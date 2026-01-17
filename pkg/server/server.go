package server

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/git"
	"github.com/charmbracelet/wish/logging"

	"github.com/nim-sam/gitport/pkg/auth"
	"github.com/nim-sam/gitport/pkg/logger"
)

const (
	dir = ".gitport"
)

func InitConfig() error {
	file, err := os.Open(filepath.Join(logger.WorkDir, logger.Conf))
	if err != nil {
		if os.IsNotExist(err) {
			logger.Logger.Warn("File not found, creating default config", "file", logger.Conf)

			var input string

			fmt.Print("Do you want the server to be public (allow guest users)? (y/n): ")
			fmt.Scan(&input)
			logger.Config.Public = (input == "y")

			fmt.Print("What is the default permission of users (none, read, write, admin): ")
			fmt.Scan(&input)
			switch input {
			case "read", "write", "admin":
				logger.Config.DefaultPerm = input
			default:
				logger.Config.DefaultPerm = "none"
			}

			file, err := os.Create(filepath.Join(logger.WorkDir, logger.Conf))
			if err != nil {
				return err
			}
			defer file.Close()

			encoder := json.NewEncoder(file)
			encoder.SetIndent("", "    ")
			err = encoder.Encode(logger.Config)
			if err != nil {
				return err
			}

			return nil
		}

		return err
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bytes, &logger.Config)
	if err != nil {
		return err
	}

	return nil
}

type Hook struct {
	repoName string
}

func (h Hook) AuthRepo(repo string, key ssh.PublicKey) git.AccessLevel {
	if repo != h.repoName {
		return git.NoAccess
	}

	userKey := key.Type() + " " + base64.StdEncoding.EncodeToString(key.Marshal())

	user, exist := auth.Data[userKey]
	if !exist {
		return git.NoAccess
	}

	switch user.Perm {
	case "read":
		return git.ReadOnlyAccess
	case "write":
		return git.ReadWriteAccess
	case "admin":
		return git.AdminAccess
	default:
		return git.NoAccess
	}
}

func (h Hook) Push(repo string, key ssh.PublicKey) {
	logger.Logger.Info("Push", "repo", repo)
}

func (h Hook) Fetch(repo string, key ssh.PublicKey) {
	logger.Logger.Info("Fetch", "repo", repo)
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
	configDir, err := os.UserConfigDir()

	if err != nil {
		logger.Logger.Error("Couldn't find user config directory", "error", err)
		return "", "", err
	}

	baseDir := filepath.Join(configDir, "gitport")
	barePath := filepath.Join(baseDir, repoName)
	logger.WorkDir = filepath.Join(barePath, dir)

	// 1. Ensure the base directory exists
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return "", "", fmt.Errorf("failed to create base directory: %w", err)
	}

	// 2. Check if the bare repository already exists
	if _, err := os.Stat(barePath); err == nil {
		// Repository already exists, return the data without cloning
		logger.Logger.Warn("Bare repo already exists, skipping clone", "path", barePath)
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
		logger.Logger.Error("Failed to create bare repo", "error", err)
		return
	}

	// Setup .gitport
	if err := os.MkdirAll(logger.WorkDir, 0755); err != nil {
		logger.Logger.Error("Failed to create .gitport", "error", err)
		return
	}

	logs := logger.InitFileLogs()
	if logs != nil {
		defer logs.Close()
	} else {
		logger.Logger.Error("Failed to initialize logs")
		return
	}

	// Initialize the config file
	err = InitConfig()
	if err != nil {
		logger.Logger.Error("Failed to initialize config", "error", err)
		return
	}

	err = auth.InitUsers()
	if err != nil {
		logger.Logger.Error("Failed to initialize users", "error", err)
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
	h := Hook{repoName}

	hostKeyPath := filepath.Join(logger.WorkDir, ".ssh", "id_ed25519")
	s, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort("0.0.0.0", port)),
		wish.WithHostKeyPath(hostKeyPath),
		wish.WithPublicKeyAuth(auth.AuthHandler),
		wish.WithMiddleware(
			git.Middleware(repoDir, h),
			gitListMiddleware(port, repoDir),
			logging.Middleware(),
		),
	)

	if err != nil {
		logger.Logger.Error("Could not start GitPort server", "error", err)
	}

	done := make(chan os.Signal, 1)

	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	logger.Logger.Info("Starting GitPort server", "repo", repoName, "URI", fullUri)

	go func() {
		// 1. Wait a moment for the server to bind the port and start listening
		time.Sleep(1 * time.Second)

		// 2. Configure the local repo to talk to the server
		configureLocalGit(fullUri)
	}()

	go func() {
		if err = s.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			logger.Logger.Error("Could not start GitPort server", "error", err)
			done <- nil
		}
	}()

	<-done

	logger.Logger.Info("Stopping GitPort server")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() { cancel() }()

	if err := s.Shutdown(ctx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
		logger.Logger.Error("Could not start GitPort server", "error", err)
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
				logger.Logger.Error("Invalid repository", "error", err)
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

// configureLocalGit sets the remote 'origin' and configures the upstream branch
func configureLocalGit(uri string) {
	logger.Logger.Info("Configuring local git remote...", "uri", uri)

	// 1. Set the remote 'origin' to our new server URI
	// We try to set-url first (in case it exists), if that fails, we add it.
	if err := exec.Command("git", "remote", "set-url", "origin", uri).Run(); err != nil {
		if err := exec.Command("git", "remote", "add", "origin", uri).Run(); err != nil {
			logger.Logger.Error("Failed to set git remote", "error", err)
			return
		}
	}

	// 2. Fetch from the remote to ensure we see the refs
	// This requires the SSH server to be up and running!
	if err := exec.Command("git", "fetch", "origin").Run(); err != nil {
		logger.Logger.Error("Failed to fetch from origin. Is the server reachable?", "error", err)
		return
	}

	// 3. Determine current branch name
	// We need to know which branch to associate with the upstream
	out, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		logger.Logger.Error("Failed to get current branch", "error", err)
		return
	}
	currentBranch := strings.TrimSpace(string(out))

	// 4. Set the upstream (tracking) information
	// This allows you to run 'git pull' and 'git push' without arguments
	upstream := fmt.Sprintf("origin/%s", currentBranch)
	if err := exec.Command("git", "branch", "--set-upstream-to="+upstream, currentBranch).Run(); err != nil {
		logger.Logger.Error("Failed to set upstream branch", "error", err)
		return
	}

	logger.Logger.Info("Git remote configured successfully. You can now use 'git push' and 'git pull'.")
}

/**
 * Starts GitPort server
 *
 * @param <host> Host address
 * @param <port> Port at which we want the ssh server to run
 * @return the full ssh server URI
 */
func Start(port string) {
	logger.InitTermLogger()

	cwd, err := os.Getwd()

	alldirs, err := os.ReadDir(cwd)
	if !ContainsFile(alldirs, ".git") {
		logger.Logger.Error("This directory doesn't contain a .git folder (Repo not initialized)")
		return
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		logger.Logger.Error("Could not create .gitport directory", "error", err)
		return
	}

	if err != nil {
		logger.Logger.Error("Could not get current directory", "error", err)
		return
	}

	gitService(port, cwd)
}
