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
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	wishbubbletea "github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"
	"github.com/charmbracelet/wish/recover"
	"github.com/go-git/go-git/v5"

	"github.com/nim-sam/gitport/internal/repository"
	"github.com/nim-sam/gitport/internal/util"
	"github.com/nim-sam/gitport/ui/ssh_tui/dashboard"
)

const (
	// Defaults used when Host and/or Port are left empty on the Server
	defaultHost = "0.0.0.0"
	defaultPort = "23234"

	// Number of most recent commits handed to the dashboard
	commitLimit = 30

	// Grace period for in-flight sessions on shutdown
	shutdownTimeout = 30 * time.Second

	// Terminal geometry assumed when a session reports no usable PTY
	fallbackWidth  = 80
	fallbackHeight = 24
)

type Server struct {
	Root      string
	Name      string
	Repo      *git.Repository
	ConfigDir string
	Host      string
	Port      string
}

// Starts GitPort server and serves it over ssh via Wish
//
// Blocks until the listener exits or the process is interrupted, at which
// point in-flight sessions are given shutdownTimeout to drain.
func (s Server) Start() error {

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

	// Advertise the routable address; "0.0.0.0" is a bind address, not
	// something a client can dial.
	dialHost := host
	if dialHost == defaultHost {
		dialHost = util.GetLocalIP()
	}

	srv, err := wish.NewServer(
		wish.WithAddress(address),
		wish.WithHostKeyPath(hostKeyPath),
		wish.WithVersion("GitPort"),
		wish.WithBanner(fmt.Sprintf("GitPort is served over ssh on %s. Type 'exit' to disconnect.", net.JoinHostPort(dialHost, port))),
		wish.WithMiddleware(
			wishbubbletea.Middleware(s.dashboardHandler),
			activeterm.Middleware(), // Bubble Tea apps usually require a PTY.
			logging.Middleware(),
			recover.Middleware(), // Keep a panicking session from taking the server down.
		),
	)
	if err != nil {
		return fmt.Errorf("could not start server: %w", err)
	}

	log.Info("Starting GitPort server",
		"listen", address,
		"connect", net.JoinHostPort(dialHost, port),
		"host_key", hostKeyPath,
	)

	// signal.NotifyContext unregisters itself on the first interrupt, restoring
	// the default behaviour so a second Ctrl-C force-quits.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			return fmt.Errorf("could not start server: %w", err)
		}
		return nil

	case <-ctx.Done():
		log.Info("Stopping GitPort server")

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
	// database state rather than a snapshot taken at startup.
	users := repository.UserRepo.Load()
	commits := s.LoadCommits(commitLimit)
	logs := repository.LogRepo.Load()

	d := dashboard.New(users, commits, logs, width, height, s.GetCommitDiff)

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
