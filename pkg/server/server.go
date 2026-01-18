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
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/git"

	"github.com/nim-sam/gitport/pkg/auth"
	"github.com/nim-sam/gitport/pkg/logger"
	"github.com/nim-sam/gitport/pkg/tui"
)

const (
	dir = ".gitport"
)

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
					// Update list for permissions
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

func InitConfig() error {
	file, err := os.Open(filepath.Join(logger.WorkDir, logger.Conf))
	if err != nil {
		if os.IsNotExist(err) {
			logger.Logger.Warn("File not found, creating default config", "file", logger.Conf)

			// Use TUI for config setup
			newConfig, err := runConfigTUI()
			if err != nil {
				// Fall back to CLI if TUI fails
				logger.Logger.Warn("TUI failed, falling back to CLI", "error", err)

				var input string
				fmt.Print("Do you want the server to be public (allow guest users)? (y/n): ")
				fmt.Scan(&input)
				newConfig.Public = (strings.ToLower(input) == "y")

				fmt.Print("What is the default permission of users (none, read, write, admin): ")
				fmt.Scan(&input)
				switch strings.ToLower(input) {
				case "read", "write", "admin":
					newConfig.DefaultPerm = input
				default:
					newConfig.DefaultPerm = "none"
				}
			}

			logger.SetConfig(newConfig)

			if err := logger.WriteJSONFile(logger.Conf, newConfig); err != nil {
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

	var newConfig logger.ConfigData
	err = json.Unmarshal(bytes, &newConfig)
	if err != nil {
		return err
	}

	logger.SetConfig(newConfig)

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

func (h Hook) Push(repo string, key ssh.PublicKey) {
	logger.Logger.Info("Push", "repo", repo)
}

func (h Hook) Fetch(repo string, key ssh.PublicKey) {
	logger.Logger.Info("Fetch", "repo", repo)
}

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

	// Ensure the base directory exists
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return "", "", fmt.Errorf("failed to create base directory: %w", err)
	}

	// Check if the bare repository already exists
	if _, err := os.Stat(barePath); err == nil {
		logger.Logger.Warn("Bare repo already exists, skipping clone", "path", barePath)
		return baseDir, repoName, nil
	}

	// Create bare repository with loading animation
	cmd := exec.Command("git", "clone", "--bare", cwd, barePath)

	// Run with loading animation
	logger.Logger.Info("Creating bare repository...")
	err = runLoadingAnimation(cmd, "git clone")
	if err != nil {
		return "", "", fmt.Errorf("failed to clone bare repository: %w", err)
	}

	return baseDir, repoName, nil
}

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

	// Ensure the host is added as admin
	err = auth.EnsureHostAdmin()
	if err != nil {
		logger.Logger.Error("Failed to ensure host admin", "error", err)
		return
	}

	// Set callback for reloading users when file changes
	logger.SetUsersReloadCallback(auth.ReloadUsers)

	// Initialize file watcher for users.json and config.json
	err = logger.InitFileWatcher()
	if err != nil {
		logger.Logger.Error("Failed to initialize file watcher", "error", err)
		return
	}
	defer logger.CloseFileWatcher()

	localIp := getLocalIP()
	fullUri := "ssh://" + net.JoinHostPort(localIp, port) + "/" + repoName

	// GitHooks implementation to allow global read write access
	h := Hook{repoName}

	hostKeyPath := filepath.Join(logger.WorkDir, ".ssh", "id_ed25519")
	s, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort("0.0.0.0", port)),
		wish.WithHostKeyPath(hostKeyPath),
		wish.WithPublicKeyAuth(auth.AuthHandler),
		wish.WithMiddleware(
			git.Middleware(repoDir, h),
			tui.Middleware(cwd),
		),
	)

	if err != nil {
		logger.Logger.Error("Could not start GitPort server", "error", err)
	}

	done := make(chan os.Signal, 1)

	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Create and run a loading animation for server startup
	go func() {
		s := spinner.New()
		s.Spinner = spinner.Dot
		s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#5000ff"))

		model := loadingModel{
			spinner:  s,
			messages: []string{"Starting GitPort server..."},
		}

		p := tea.NewProgram(model)

		// Send updates to the loading animation
		go func() {
			time.Sleep(1 * time.Second)
			p.Send(loadingMsg(fmt.Sprintf("Repository: %s", repoName)))
			p.Send(loadingMsg(fmt.Sprintf("Server URI: %s", fullUri)))
			p.Send(loadingMsg("Configuring local git remote..."))

			configureLocalGit(fullUri)

			time.Sleep(2 * time.Second)
			p.Send("done")
		}()

		p.Run()
	}()

	time.Sleep(3 * time.Second) // Give time for animation to show
	logger.Logger.Info("Starting GitPort server", "repo", repoName, "URI", fullUri)

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
		logger.Logger.Error("Could not stop GitPort server", "error", err)
	}
}

func ContainsFile(d []os.DirEntry, name string) bool {
	for _, entry := range d {
		if entry.Name() == name {
			return true
		}
	}
	return false
}

func configureLocalGit(uri string) {
	logger.Logger.Info("Configuring local git remote...", "uri", uri)

	// Set the remote 'origin'
	if err := exec.Command("git", "remote", "set-url", "origin", uri).Run(); err != nil {
		if err := exec.Command("git", "remote", "add", "origin", uri).Run(); err != nil {
			logger.Logger.Error("Failed to set git remote", "error", err)
			return
		}
	}

	// Fetch from the remote
	if err := exec.Command("git", "fetch", "origin").Run(); err != nil {
		logger.Logger.Warn("Could not fetch from origin. If this is a new repo, this is normal.", "error", err)
	}

	// Determine current branch name
	var currentBranch string
	out, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()

	if err != nil {
		// HEAD is invalid (empty repo)
		defaultBranch, configErr := exec.Command("git", "config", "--get", "init.defaultBranch").Output()
		if configErr == nil && len(defaultBranch) > 0 {
			currentBranch = strings.TrimSpace(string(defaultBranch))
		} else {
			currentBranch = "master"
		}
		logger.Logger.Info("Empty repository detected. Future commits will track", "branch", currentBranch)
	} else {
		currentBranch = strings.TrimSpace(string(out))
	}

	// Set the upstream (tracking) information
	upstream := fmt.Sprintf("origin/%s", currentBranch)
	if err := exec.Command("git", "branch", "--set-upstream-to="+upstream, currentBranch).Run(); err != nil {
		logger.Logger.Info("Remote branch not found yet. To push and link, run:",
			"command", fmt.Sprintf("git push -u origin %s", currentBranch))
	} else {
		logger.Logger.Info("Git remote configured successfully. Tracking", "upstream", upstream)
	}
}

func Start(port string) {
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
