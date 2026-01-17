package server

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/git"
	"github.com/charmbracelet/wish/logging"
)

type app struct {
	access git.AccessLevel
}

func (a app) AuthRepo(string, ssh.PublicKey) git.AccessLevel {
	return a.access
}

func (a app) Push(repo string, _ ssh.PublicKey) {
	log.Info("push", "repo", repo)
}

func (a app) Fetch(repo string, _ ssh.PublicKey) {
	log.Info("fetch", "repo", repo)
}

/*
 * Returns host IP address
 */
func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "localhost"
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "localhost"
}

/**
 * Creates a bare repository of the repository located at <cwd>
 *
 * @param <cwd> Current Working Directory
 * @return git server directory, repository name, error
 */
func createBareRepo(cwd string) (string, string, error) {
	repoName := filepath.Base(cwd) + ".git"
	baseDir := filepath.Join(os.TempDir(), "gitport")
	barePath := filepath.Join(baseDir, repoName)

	// 1. Ensure the base directory exists
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return "", "", fmt.Errorf("failed to create base directory: %w", err)
	}

	// 2. Check if the bare repository already exists
	if _, err := os.Stat(barePath); err == nil {
		// Repository already exists, return the data without cloning
		log.Debug("Bare repo already exists, skipping clone", "path", barePath)
		return baseDir, repoName, nil
	}

	// 3. If it doesn't exist, create it via bare clone
	cmd := exec.Command("git", "clone", "--bare", cwd, barePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", "", fmt.Errorf("failed to clone bare repository: %w", err)
	}

	return baseDir, repoName, nil
}

/*
 * Creates SSH server and enables git operations
 *
 * @param <port> Port to bind to
 * @param <cwd> Current Working Directory i.e. Location of the Git repository
 */
func gitService(port string, cwd string) {

	repoDir, repoName, err := createBareRepo(cwd)
	if err != nil {
		log.Error("Failed to create bare repo", "error", err)
		return
	}

	localIp := getLocalIP()
	fullUri := "ssh://" + net.JoinHostPort(localIp, port) + "/" + repoName

	// 1. Add the local server as a remote (e.g., named 'origin' or 'gitport')
	//exec.Command("git", "remote", "add", "origin", fullUri).Run()

	// 2. Set the upstream tracking
	// This tells the local branch to track the remote branch
	//exec.Command("git", "push", "--set-upstream", "origin", "master").Run()

	// GitHooks implementation to allow global read write access
	a := app{
		git.ReadWriteAccess,
	}

	s, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort("0.0.0.0", port)),
		wish.WithHostKeyPath(".ssh/id_ed25519"),

		// // Accept all public keys
		// ssh.PublicKeyAuth(
		// 	func(ssh.Context, ssh.PublicKey) bool {
		// 		return true
		// 	},
		// ),

		// // Do not accept password auth
		// ssh.PasswordAuth(
		// 	func(ssh.Context, string) bool {
		// 		return false
		// 	},
		// ),

		wish.WithMiddleware(
			git.Middleware(repoDir, a),
			gitListMiddleware(port, repoDir),
			logging.Middleware(),
		),
	)

	if err != nil {
		log.Error("Could not start GitPort server", "error", err)
	}

	done := make(chan os.Signal, 1)

	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	log.Info("Starting GitPort server", "repo", repoName, "URI", fullUri)

	go func() {
		if err = s.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			log.Error("Could not start GitPort server", "error", err)
			done <- nil
		}
	}()

	<-done

	log.Info("Stopping GitPort server")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() { cancel() }()

	if err := s.Shutdown(ctx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
		log.Error("Could not start GitPort server", "error", err)
	}
}

func gitListMiddleware(port string, repoDir string) wish.Middleware {
	return func(next ssh.Handler) ssh.Handler {
		return func(sess ssh.Session) {

			localIp := getLocalIP()

			// Git will have a command included so only run this if there are no
			// commands passed to ssh.
			if len(sess.Command()) != 0 {
				next(sess)
				return
			}

			dest, err := os.ReadDir(repoDir)
			if err != nil && err != fs.ErrNotExist {
				log.Error("Invalid repository", "error", err)
			}
			if len(dest) > 0 {
				fmt.Fprintf(sess, "\n### Repo Menu ###\n\n")
			}
			for _, dir := range dest {
				wish.Println(sess, fmt.Sprintf("• %s - ", dir.Name()))
				wish.Println(sess, fmt.Sprintf("git clone ssh://%s/%s", net.JoinHostPort(localIp, port), dir.Name()))
			}
			wish.Printf(sess, "\n\n### Add some repos! ###\n\n")
			wish.Printf(sess, "> cd some_repo\n")
			wish.Printf(sess, "> git remote add wish_test ssh://%s/some_repo\n", net.JoinHostPort(localIp, port))
			wish.Printf(sess, "> git push wish_test\n\n\n")
			next(sess)
		}
	}
}

/**
 * Helper function to check if file with name <name> exists within the DirEntry array <d>
 *
 * @param <d> DirEntry array
 * @param <name> name of the target file
 * @return true if the file is present, false otherwise
 */
func ContainsFile(d []os.DirEntry, name string) bool {
	for _, entry := range d {
		if entry.Name() == name {
			return true
		}
	}
	return false
}

/**
 * Starts GitPort server
 *
 * @param <host> Host address
 * @param <port> Port at which we want the ssh server to run
 * @return the full ssh server URI
 */
func Start(port string) {

	cwd, err := os.Getwd()

	alldirs, err := os.ReadDir(cwd)
	if !ContainsFile(alldirs, ".git") {
		log.Error("This directory doesn't contain a .git folder (Repo not initialized)")
		return
	}

	if err != nil {
		log.Error("Could not get current directory", "error", err)
		return
	}

	gitService(port, cwd)
}
