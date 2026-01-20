package server

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/git"

	"github.com/nim-sam/gitport/pkg/auth"
	"github.com/nim-sam/gitport/pkg/logger"
	"github.com/nim-sam/gitport/pkg/tui"
)

const (
	gpConfig = ".gitport"
)

// GitPortServer represents the main server instance
type GpServer struct {
	Port      string
	RepoDir   string
	RepoName  string
	configDir string
}

// TUI Styles
var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("#5000ff"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	quitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4)
)

// Setup TUI Models
type configItem string

func (i configItem) FilterValue() string { return string(i) }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(configItem)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

type configModel struct {
	list        list.Model
	choice      string
	quitting    bool
	step        int
	public      bool
	defaultPerm string
	config      logger.ConfigData
}

func (m configModel) Init() tea.Cmd {
	return nil
}

func (m configModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			i, ok := m.list.SelectedItem().(configItem)
			if ok {
				m.choice = string(i)

				switch m.step {
				case 0: // Public selection
					m.public = (m.choice == "Yes")
					m.step = 1
					items := []list.Item{
						configItem("none"),
						configItem("read"),
						configItem("write"),
						configItem("admin"),
					}
					m.list.Title = "What is the default permission of users?"
					m.list.SetItems(items)
					m.choice = ""
					return m, nil

				case 1: // Permission selection
					m.defaultPerm = m.choice
					m.config.Public = m.public
					m.config.DefaultPerm = m.defaultPerm
					return m, tea.Quit
				}
			}
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m configModel) View() string {
	if m.choice != "" && m.step == 1 {
		return quitTextStyle.Render(fmt.Sprintf("Config saved!\nPublic: %v\nDefault Permission: %s",
			m.public, m.defaultPerm))
	}
	if m.quitting {
		return quitTextStyle.Render("Setup cancelled.")
	}
	return "\n" + m.list.View()
}

// runConfigTUI presents a TUI for configuring server settings
func runConfigTUI() (logger.ConfigData, error) {
	items := []list.Item{
		configItem("Yes"),
		configItem("No"),
	}

	const defaultWidth = 20
	const listHeight = 4

	l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
	l.Title = "Do you want the server to be public (allow guest users)?"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	m := configModel{
		list:   l,
		step:   0,
		config: logger.ConfigData{},
	}

	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return logger.ConfigData{}, err
	}

	if m, ok := finalModel.(configModel); ok {
		return m.config, nil
	}

	return logger.ConfigData{}, fmt.Errorf("failed to get config from TUI")
}

// Loading Animation Model
type loadingMsg string

type loadingModel struct {
	spinner  spinner.Model
	messages []string
	done     bool
	err      error
	width    int
	height   int
}

func (m loadingModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m loadingModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m, tea.Quit
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case loadingMsg:
		m.messages = append(m.messages, string(msg))
		return m, nil
	case error:
		m.err = msg
		m.done = true
		return m, tea.Quit
	case string:
		if msg == "done" {
			m.done = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m loadingModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("\n Error: %v\n Press any key to exit\n", m.err)
	}

	var output strings.Builder

	// Header
	output.WriteString("\n")
	output.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("#5000ff")).
		Bold(true).
		Padding(0, 1).
		Render("GitPort Server Setup"))
	output.WriteString("\n\n")

	// Spinner and status
	status := "Setting up server..."
	if m.done {
		status = "Setup complete!"
	}
	output.WriteString(fmt.Sprintf(" %s %s\n\n", m.spinner.View(), status))

	// Messages log
	output.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("#5000ff")).
		Padding(0, 1).
		Render("Progress:"))
	output.WriteString("\n")

	// Show last 5 messages
	start := 0
	if len(m.messages) > 5 {
		start = len(m.messages) - 5
	}
	for i := start; i < len(m.messages); i++ {
		output.WriteString(fmt.Sprintf("  • %s\n", m.messages[i]))
	}

	if !m.done {
		output.WriteString("\n" + lipgloss.NewStyle().
			Foreground(lipgloss.Color("#5000ff")).
			Render("Press any key to skip animation"))
	}

	return output.String()
}

/*
 * runLoadingAnimation starts loading animation for a command execution
 *
 * @param cmd command to execute
 * @param taskName name of the task for display
 * @return error if any
 */
func runLoadingAnimation(cmd *exec.Cmd, taskName string) error {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#5000ff"))

	initialModel := loadingModel{
		spinner:  s,
		messages: []string{fmt.Sprintf("Starting %s...", taskName)},
	}

	p := tea.NewProgram(initialModel)

	// Create pipes for stdout and stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	// Start command
	if err := cmd.Start(); err != nil {
		p.Send(err)
		return err
	}

	// Goroutine to read output
	go func() {
		scanner := bufio.NewScanner(io.MultiReader(stdout, stderr))
		for scanner.Scan() {
			line := scanner.Text()
			p.Send(loadingMsg(line))
		}

		// Wait for command to complete
		if err := cmd.Wait(); err != nil {
			p.Send(err)
			return
		}

		p.Send("done")
	}()

	// Run the TUI
	_, err = p.Run()
	return err
}

// Hook implements Git hook callbacks for authentication and access control
type Hook struct {
	repoName string
}

// AuthRepo determines the access level for a user based on their key and repository
func (h Hook) AuthRepo(repo string, key ssh.PublicKey) git.AccessLevel {
	if repo != h.repoName {
		return git.NoAccess
	}

	userKey := key.Type() + " " + base64.StdEncoding.EncodeToString(key.Marshal())
	user, exist := auth.GetUserByKey(userKey)
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

// Push logs push operations to the repository
func (h Hook) Push(repo string, key ssh.PublicKey) {
	logger.Logger.Info("Push", "repo", repo)
}

// Fetch logs fetch operations from the repository
func (h Hook) Fetch(repo string, key ssh.PublicKey) {
	logger.Logger.Info("Fetch", "repo", repo)
}

// getLocalIP returns the local IP address of the machine
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

// initBareRepo creates a bare Git repository for serving
func initBareRepo(cwd string) (string, string, string, error) {
	repoName := filepath.Base(cwd) + ".git"

	configDir, err := os.UserConfigDir()
	if err != nil {
		logger.Logger.Error("Couldn't find user config directory", "error", err)
		return "", "", "", err
	}

	baseDir := filepath.Join(configDir, "gitport")
	barePath := filepath.Join(baseDir, repoName)

	gpConf := filepath.Join(barePath, ".gitport")

	if _, err := os.Open(gpConf); err != nil {
		return "", "", "", fmt.Errorf("Cannot start an uninitialized server. Run `git init` before starting.")
	}

	// Ensure the base directory exists
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return "", "", "", fmt.Errorf("failed to create base directory: %w", err)
	}

	// Check if the bare repository already exists
	if _, err := os.Stat(barePath); err == nil {
		logger.Logger.Warn("Bare repo already exists, skipping clone", "path", barePath)
		return baseDir, repoName, gpConf, nil
	}

	// Create bare repository with loading animation
	err = exec.Command("git", "clone", "--bare", cwd, barePath).Run()
	logger.Logger.Info("Creating bare repository...")
	//err = runLoadingAnimation(cmd, "git clone")
	if err != nil {
		return "", "", "", fmt.Errorf("failed to clone bare repository: %w", err)
	}

	return baseDir, repoName, configDir, nil
}

// InitConfig initializes or loads server configuration
func (s GpServer) initConfig() error {

	logger.ConfigDir = s.configDir

	// Setup .gitport directory
	if err := os.MkdirAll(s.configDir, 0755); err != nil {
		return fmt.Errorf("failed to create .gitport directory: %w", err)
	}

	// Checks if a default config file exists within the .gitport
	// folder (using config.json in this case)
	configFilePath := filepath.Join(s.configDir, logger.Conf)

	// Check if .gitport exists and has content
	file, err := os.Open(configFilePath)

	if err != nil {
		if os.IsNotExist(err) {
			return createDefaultConfig()
		}
		return err
	} else {
		println("GitPort server already initialized. Run\n\n\tgitport start <port>\n\nto start GitPort server for %s", s.RepoName)
	}

	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	var newConfig logger.ConfigData
	err = json.Unmarshal(bytes, &newConfig)
	if err != nil {
		return err
	}

	logger.SetConfig(newConfig)
	return nil
}

// createDefaultConfig creates a new configuration file with user input
func createDefaultConfig() error {
	logger.Logger.Warn("File not found, creating default config", "file", logger.Conf)

	// Use TUI for config setup
	newConfig, err := runConfigTUI()
	if err != nil {
		logger.Logger.Warn("TUI failed, falling back to CLI", "error", err)
		newConfig = getCLIConfig()
	}

	logger.SetConfig(newConfig)
	return logger.WriteJSONFile(logger.Conf, newConfig)
}

// getCLIConfig prompts user for configuration via CLI
func getCLIConfig() logger.ConfigData {
	var config logger.ConfigData
	var input string

	fmt.Print("Do you want the server to be public (allow guest users)? (y/n): ")
	fmt.Scan(&input)
	config.Public = (strings.ToLower(input) == "y")

	fmt.Print("What is the default permission of users (none, read, write, admin): ")
	fmt.Scan(&input)
	switch strings.ToLower(input) {
	case "read", "write", "admin":
		config.DefaultPerm = input
	default:
		config.DefaultPerm = "none"
	}

	return config
}

// sets up all server components (logs, auth, file watcher)
func (s *GpServer) initGitPortServer() error {

	logger.ConfigDir = s.configDir

	// Initialize file logging
	logs := logger.Logger.InitFileLogs(s.configDir)
	if logs == nil {
		return fmt.Errorf("failed to initialize logs")
	}
	defer logs.Close()

	// Initialize users and authentication
	if err := auth.InitUsers(); err != nil {
		return fmt.Errorf("failed to initialize users: %w", err)
	}

	if err := auth.EnsureHostAdmin(); err != nil {
		return fmt.Errorf("failed to ensure host admin: %w", err)
	}

	// Set up file change callbacks
	logger.SetUsersReloadCallback(auth.ReloadUsers)

	// Initialize file watcher
	if err := logger.InitFileWatcher(); err != nil {
		return fmt.Errorf("failed to initialize file watcher: %w", err)
	}

	// if err := s.initConfig(); err != nil {
	// 	return err
	// }

	return nil
}

// startGitPortServer starts the SSH server with Git middleware
func (s GpServer) startGitPortServer() error {
	localIP := getLocalIP()
	fullURI := "ssh://" + net.JoinHostPort(localIP, s.Port) + "/" + s.RepoName

	hook := Hook{repoName: s.RepoName}
	hostKeyPath := filepath.Join(s.configDir, ".ssh", "id_ed25519")

	server, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort("0.0.0.0", s.Port)),
		wish.WithHostKeyPath(hostKeyPath),
		wish.WithPublicKeyAuth(auth.AuthHandler),
		wish.WithMiddleware(
			git.Middleware(s.RepoDir, hook),
			tui.Middleware("."),
		),
	)

	if err != nil {
		return fmt.Errorf("could not create server: %w", err)
	}

	// Start server with loading animation
	showServerStartupAnimation(s.RepoName, fullURI)

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err = server.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			logger.Logger.Error("Could not start GitPort server", "error", err)
			done <- nil
		}
	}()

	<-done
	return shutdownServer(server)
}

// showServerStartupAnimation displays a loading animation during server startup
func showServerStartupAnimation(repoName, fullURI string) {
	go func() {
		s := spinner.New()
		s.Spinner = spinner.Dot
		s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#5000ff"))

		model := loadingModel{
			spinner:  s,
			messages: []string{"Starting GitPort server..."},
		}

		p := tea.NewProgram(model)

		go func() {
			time.Sleep(1 * time.Second)
			p.Send(loadingMsg(fmt.Sprintf("Repository: %s", repoName)))
			p.Send(loadingMsg(fmt.Sprintf("Server URI: %s", fullURI)))
			p.Send(loadingMsg("Configuring local git remote..."))
			configureLocalGit(fullURI)
			time.Sleep(2 * time.Second)
			p.Send("done")
		}()

		p.Run()
	}()

	time.Sleep(3 * time.Second) // Give time for animation to show
	logger.Logger.Info("Starting GitPort server", "repo", repoName, "URI", fullURI)
}

// shutdownServer gracefully shuts down the server
func shutdownServer(server *ssh.Server) error {
	logger.Logger.Info("Stopping GitPort server")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
		return fmt.Errorf("could not stop GitPort server: %w", err)
	}
	return nil
}

// ContainsFile checks if a file exists in a directory
func ContainsFile(d []os.DirEntry, name string) bool {
	for _, entry := range d {
		if entry.Name() == name {
			return true
		}
	}
	return false
}

// configureLocalGit sets up the local Git repository to use the server as remote
func configureLocalGit(uri string) {
	logger.Logger.Info("Configuring local git remote...", "uri", uri)

	// Set the remote 'origin'
	if err := exec.Command("git", "remote", "set-url", "origin", uri).Run(); err != nil {
		if err := exec.Command("git", "remote", "add", "origin", uri).Run(); err != nil {
			logger.Logger.Error("Failed to set git remote", "error", err)
			return
		}
	}

	// Fetch from remote
	if err := exec.Command("git", "fetch", "origin").Run(); err != nil {
		logger.Logger.Warn("Could not fetch from origin. If this is a new repo, this is normal.", "error", err)
	}

	setUpstreamBranch()
}

// setUpstreamBranch configures the upstream tracking branch for the current repository
func setUpstreamBranch() {
	currentBranch := getCurrentBranch()
	upstream := fmt.Sprintf("origin/%s", currentBranch)

	if err := exec.Command("git", "branch", "--set-upstream-to="+upstream, currentBranch).Run(); err != nil {
		logger.Logger.Info("Remote branch not found yet. To push and link, run:",
			"command", fmt.Sprintf("git push -u origin %s", currentBranch))
	} else {
		logger.Logger.Info("Git remote configured successfully. Tracking", "upstream", upstream)
	}
}

// getCurrentBranch determines the current Git branch name
func getCurrentBranch() string {
	out, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err == nil && len(out) > 0 {
		return strings.TrimSpace(string(out))
	}

	// HEAD is invalid (empty repo) - get default branch
	defaultBranch, configErr := exec.Command("git", "config", "--get", "init.defaultBranch").Output()
	if configErr == nil && len(defaultBranch) > 0 {
		branch := strings.TrimSpace(string(defaultBranch))
		logger.Logger.Info("Empty repository detected. Future commits will track", "branch", branch)
		return branch
	}

	return "master"
}

// Initialize GitPort server
func Init() {
	cwd, _ := os.Getwd()
	dirs, _ := os.ReadDir(cwd)

	if !ContainsFile(dirs, ".git") {
		log.Error("This directory doesn't contain a .git folder (Repo not initialized)")
		return
	}

	repoDir, repoName, _, err := initBareRepo(cwd)
	if err != nil {
		log.Error("Failed to create bare repo", "error", err)
		return
	}

	// Initialize server struct
	userConf, _ := os.UserConfigDir()
	gpConf := filepath.Join(userConf, "gitport", repoName, ".gitport")

	server := GpServer{
		RepoName:  repoName,
		RepoDir:   repoDir,
		configDir: gpConf,
	}

	if err := server.initConfig(); err != nil {
		log.Error("Failed to initialize config", "error", err)
		return
	}

	if err := server.initGitPortServer(); err != nil {
		log.Error("Failed to initialize server components", "error", err)
		return
	}
}

// Start GitPort server on the specified port
func Start(port string) {

	cwd, _ := os.Getwd()

	if dirs, _ := os.ReadDir(cwd); !ContainsFile(dirs, ".git") {
		log.Error("This directory doesn't contain a .git folder (Repo not initialized)")
		return
	}

	repoDir, repoName, gpConf, err := initBareRepo(cwd)

	if err != nil {
		log.Error("Failed to fetch bare repo", "error", err)
		return
	}

	// Initialize server struct

	server := GpServer{
		RepoName:  repoName,
		RepoDir:   repoDir,
		Port:      port,
		configDir: gpConf,
	}

	if err := server.initGitPortServer(); err != nil {
		logger.Logger.Error("Failed to initialize server components", "error", err)
		return
	}
	defer logger.CloseFileWatcher()

	if err := server.startGitPortServer(); err != nil {
		logger.Logger.Error("Server error", "error", err)
	}
}
