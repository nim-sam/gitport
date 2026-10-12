package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	wishbubbletea "github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"
	"github.com/charmbracelet/wish/recover"
	"github.com/go-git/go-git/v5"

	"github.com/nim-sam/gitport/internal/config"
	"github.com/nim-sam/gitport/internal/logger"
	"github.com/nim-sam/gitport/internal/mdns"
	"github.com/nim-sam/gitport/internal/repository"
	"github.com/nim-sam/gitport/internal/util"
	"github.com/nim-sam/gitport/ui/ssh_tui/dashboard"
)

const (
	// Defaults used when Host and/or Port are left empty on the Server
	defaultHost = "0.0.0.0"
	defaultPort = "22"

	// Number of most recent commits handed to the dashboard
	commitLimit = 30

	// Grace period for in-flight sessions on shutdown
	shutdownTimeout = 30 * time.Second

	// Terminal geometry assumed when a session reports no usable PTY
	fallbackWidth  = 80
	fallbackHeight = 24
)

type Server struct {
	Root string
	Name string
	Repo *git.Repository
	Host string
	Port string

	// Ready, when non-nil, is closed once the TCP listener is bound and the
	// server is accepting connections. It lets a caller (such as the start TUI)
	// tell "listening" apart from "still starting". Run closes it at most once,
	// so a Server is intended to be served a single time.
	Ready chan struct{}
}

// New builds a Server for the repository rooted at root from persisted
// configuration.
//
// Opening git is best-effort: a repository without commits, or a directory
// with no .git at all, still yields a usable server, it simply has no history
// to display. The display name falls back to the repository directory's name
// so an unset config still announces something meaningful over mDNS.
func New(root string, cfg config.Config) Server {
	repo, err := git.PlainOpen(root)
	if err != nil {
		repo = nil
	}

	name := cfg.DisplayName
	if name == "" {
		name = filepath.Base(root)
	}

	return Server{
		Root: root,
		Repo: repo,
		Name: name,
		Host: cfg.Host,
		Port: cfg.Port,
	}
}

// Start runs the server until the process receives SIGINT/SIGTERM. It is a
// convenience wrapper around Run for callers that do not manage a context.
func (s Server) Start() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	return s.Run(ctx)
}

// Run serves GitPort over ssh via Wish until ctx is cancelled or the listener
// fails. The server is also announced on the local network via mDNS as
// <repo>-gitport.local.
//
// Blocks until the listener exits or ctx is done, at which point in-flight
// sessions are given shutdownTimeout to drain. Ready, when set, is closed after
// the listener binds successfully.
func (s Server) Run(ctx context.Context) error {

	host, port := s.Host, s.Port
	if host == "" {
		host = defaultHost
	}
	if port == "" {
		port = defaultPort
	}

	hostKeyPath, err := s.hostKeyPath()
	if err != nil {
		return err
	}

	address := net.JoinHostPort(host, port)

	// mDNS needs a concrete IP to publish. "0.0.0.0" is a bind address, not
	// something a client can dial, and a hostname can't be announced as an
	// A record, so fall back to the detected LAN address in both cases.
	announceIP := host
	if announceIP == defaultHost || net.ParseIP(announceIP) == nil {
		announceIP = util.GetLocalIP()
	}

	// Announce <repo>-gitport.local. Failure is non-fatal: the server still
	// works by IP. On error Announce returns a no-op cleanup, so the defer
	// is safe either way.
	dialHost, stopMDNS, err := mdns.Announce(s.Name, port, announceIP)
	if err != nil {
		logger.Warn("mDNS announce failed, serving by IP only", "err", err)
		dialHost = announceIP
	}
	defer stopMDNS()

	connect := "ssh " + dialHost
	if port != "22" {
		connect = fmt.Sprintf("ssh -p %s %s", port, dialHost)
	}

	srv, err := wish.NewServer(
		wish.WithAddress(address),
		wish.WithHostKeyPath(hostKeyPath),
		wish.WithVersion("GitPort"),
		wish.WithPublicKeyAuth(s.authorize),
		wish.WithBanner(fmt.Sprintf("GitPort is served over ssh. Connect with: %s. Type 'exit' to disconnect.", connect)),
		wish.WithMiddleware(
			wishbubbletea.Middleware(s.dashboardHandler),
			activeterm.Middleware(), // Bubble Tea apps usually require a PTY.
			s.gitMiddleware(),       // Serve git over the same endpoint; non-git sessions fall through to the TUI.
			logging.Middleware(),
			recover.Middleware(), // Keep a panicking session from taking the server down.
		),
	)
	if err != nil {
		return fmt.Errorf("could not start server: %w", err)
	}

	// Bind explicitly rather than through ListenAndServe so callers can observe
	// the moment the listener is up.
	ln, err := net.Listen("tcp", address)
	if err != nil {
		if errors.Is(err, syscall.EACCES) {
			return fmt.Errorf("cannot bind %s: permission denied. Grant CAP_NET_BIND_SERVICE "+
				"(systemd AmbientCapabilities or `setcap 'cap_net_bind_service=+ep' <binary>`), "+
				"or redirect 22 to a high port: %w", address, err)
		}
		if errors.Is(err, syscall.EADDRINUSE) {
			return fmt.Errorf("%s is already in use (is the system sshd still on port %s?): %w", address, port, err)
		}
		return fmt.Errorf("could not start server: %w", err)
	}

	if s.Ready != nil {
		close(s.Ready)
	}

	logger.Info("Starting GitPort server",
		"listen", address,
		"connect", connect,
		"clone", s.CloneAddress(),
		"host_key", hostKeyPath,
	)

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Serve(ln)
	}()

	select {
	case err := <-errCh:
		if err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			return fmt.Errorf("could not start server: %w", err)
		}
		return nil

	case <-ctx.Done():

		logger.Info("Stopping GitPort server")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			return fmt.Errorf("could not stop server: %w", err)
		}
		return nil
	}
}

// dashboardHandler wires a freshly loaded dashboard to every inbound ssh
// session. A new model per session is required: tea.Model is a value type and
// bubbletea mutates no shared state across programs.
//
// The session's PTY dimensions are passed down explicitly. Bubble Tea cannot
// detect the size of a remote terminal, and wish's middleware only emits
// tea.WindowSizeMsg for resizes that happen *after* the program starts, so
// without this the first frame would be laid out at zero size.
func (s Server) dashboardHandler(sess ssh.Session) (tea.Model, []tea.ProgramOption) {

	pty, _, ok := sess.Pty()

	width, height := fallbackWidth, fallbackHeight
	if ok && pty.Window.Width > 0 && pty.Window.Height > 0 {
		width, height = pty.Window.Width, pty.Window.Height
	}

	// Reload per session so the dashboard always reflects current repo and
	// database state rather than a snapshot taken at startup. Commits (with
	// their diffs) are read from the database populated by InitCommits.
	users := repository.UserRepo.Load()
	commits := repository.LoadCommitsNewestFirst()
	logs := repository.LogRepo.Load()

	// The authorized user drives what the dashboard lets this session do. When
	// the key cannot be resolved, the zero user is passed so the dashboard fails
	// closed to read-only.
	current, ok := s.authenticatedUser(sess)
	logger.Info("ssh session opened", "user", userName(current, ok))
	if ok {
		recordLogin(current)
	}

	d := dashboard.New(users, commits, logs, width, height, current)

	return d, []tea.ProgramOption{tea.WithAltScreen()}
}

// hostKeyPath returns the path to the server's SSH host key. The containing
// directory is created if absent, since wish only generates the key itself and
// fails outright when the parent directory does not exist.
func (s Server) hostKeyPath() (string, error) {
	dir := filepath.Join(s.Root, ".gitport")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", fmt.Errorf("could not create %s: %w", dir, err)
	}
	return filepath.Join(dir, "id_ed25519"), nil
}
