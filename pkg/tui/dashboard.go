package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/nim-sam/gitport/pkg/auth"
	"github.com/nim-sam/gitport/pkg/logger"
)

type formState int

const (
	stateNormal formState = iota
	stateCreating
	stateDeleting
)

type dashboardModel struct {
	userList     list.Model
	state        formState
	selectedUser string
	selectedKey  string
	width        int
	height       int

	// Form inputs for creating new user
	nameInput  textinput.Model
	keyInput   textinput.Model
	permValue  string // Current permission value (cycles through options)
	nameActive bool   // true if name field is focused, false if key field is focused
}

type userItem struct {
	key  string
	name string
	perm string
}

func (i userItem) FilterValue() string { return i.name }
func (i userItem) Title() string       { return i.name }
func (i userItem) Description() string { return fmt.Sprintf("Permission: %s", i.perm) }

func newDashboard() dashboardModel {
	items := loadUsers()

	l := list.New(items, list.NewDefaultDelegate(), 40, 14)
	l.Title = "Users"
	l.SetShowHelp(false)

	// Create form inputs
	nameInput := textinput.New()
	nameInput.Placeholder = "username"
	nameInput.CharLimit = 50

	keyInput := textinput.New()
	keyInput.Placeholder = "ssh-ed25519 AAAA..."
	keyInput.CharLimit = 500

	return dashboardModel{
		userList:   l,
		state:      stateNormal,
		nameInput:  nameInput,
		keyInput:   keyInput,
		permValue:  "none",
		nameActive: true,
	}
}

func loadUsers() []list.Item {
	var items []list.Item
	users := auth.GetAllUsers()

	for key, user := range users {
		items = append(items, userItem{
			key:  key,
			name: user.Name,
			perm: user.Perm,
		})
	}
	return items
}

func (m dashboardModel) Init() tea.Cmd {
	return nil
}

func (m dashboardModel) Update(msg tea.Msg) (dashboardModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	// dashboard.go - Update method for WindowSizeMsg
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Calculate heights properly
		totalHeight := m.height
		helpHeight := 1 // Help text at bottom

		// List takes full height minus help
		listHeight := totalHeight - helpHeight
		listWidth := int(float64(m.width) * 0.6)

		m.userList.SetSize(listWidth, listHeight)

	case tea.KeyMsg:
		if m.state == stateCreating {
			return m.handleCreatingKeys(msg)
		} else if m.state == stateDeleting {
			return m.handleDeletingKeys(msg)
		}

		switch msg.String() {
		case "n":
			m.state = stateCreating
			m.nameInput.SetValue("")
			m.keyInput.SetValue("")
			m.permValue = "none"
			m.nameActive = true
			m.nameInput.Focus()
			m.keyInput.Blur()
			return m, textinput.Blink

		case "d":
			if item, ok := m.userList.SelectedItem().(userItem); ok {
				m.selectedUser = item.name
				m.selectedKey = item.key
				m.state = stateDeleting
			}

		case "p":
			if item, ok := m.userList.SelectedItem().(userItem); ok {
				cycleUserPerm(item.key)
				m.userList.SetItems(loadUsers())
			}

		case "P":
			cycleDefaultPerm()

		case "t":
			togglePublic()
		}
	}

	var cmd tea.Cmd
	m.userList, cmd = m.userList.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m dashboardModel) handleCreatingKeys(msg tea.KeyMsg) (dashboardModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "esc":
		m.state = stateNormal
		return m, nil

	case "enter":
		// Always try to create when enter is pressed if we have valid data
		name := strings.TrimSpace(m.nameInput.Value())
		key := strings.TrimSpace(m.keyInput.Value())

		if name != "" && key != "" {
			createUser(key, name, m.permValue)
			m.userList.SetItems(loadUsers())
			m.state = stateNormal
			return m, nil
		}
		// If on name and it's filled, move to key
		if m.nameActive && name != "" {
			m.nameActive = false
			m.nameInput.Blur()
			cmd = m.keyInput.Focus()
			return m, cmd
		}

	case "p":
		// Cycle permission forward
		m.permValue = cyclePermValue(m.permValue, 1)
		return m, nil

	case "down", "ctrl+j":
		// Move to next field (name -> key -> name)
		if m.nameActive {
			m.nameActive = false
			m.nameInput.Blur()
			cmd = m.keyInput.Focus()
		} else {
			m.nameActive = true
			m.keyInput.Blur()
			cmd = m.nameInput.Focus()
		}
		return m, cmd

	case "up", "ctrl+k":
		// Move to previous field (key -> name -> key)
		if m.nameActive {
			m.nameActive = false
			m.nameInput.Blur()
			cmd = m.keyInput.Focus()
		} else {
			m.nameActive = true
			m.keyInput.Blur()
			cmd = m.nameInput.Focus()
		}
		return m, cmd
	}

	// Update the focused input
	if m.nameActive {
		m.nameInput, cmd = m.nameInput.Update(msg)
	} else {
		m.keyInput, cmd = m.keyInput.Update(msg)
	}

	return m, cmd
}

func (m dashboardModel) handleDeletingKeys(msg tea.KeyMsg) (dashboardModel, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.state = stateNormal
		m.selectedUser = ""
		m.selectedKey = ""

	case "d":
		if m.selectedKey != "" {
			deleteUser(m.selectedKey)
			m.state = stateNormal
			m.selectedUser = ""
			m.selectedKey = ""
			// Reload and reset list
			newItems := loadUsers()
			m.userList.SetItems(newItems)
			// Reset selection to first item if list is not empty
			if len(newItems) > 0 {
				m.userList.Select(0)
			}
		}
	}
	return m, nil
}

func (m dashboardModel) View() string {
	if m.state == stateCreating {
		return m.renderCreateForm()
	}
	if m.state == stateDeleting {
		return m.renderDeleteConfirm()
	}

	// Config section - match commit history colors
	isPublic := logger.GetConfigPublic()
	defaultPerm := logger.GetConfigDefaultPerm()

	listWidth := int(float64(m.width) * 0.6)
	configWidth := m.width - listWidth

	listStyle := lipgloss.NewStyle().Width(listWidth)

	configStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#5000ff")).
		Padding(0, 1).
		Width(configWidth)

	publicStr := "🔒 Private"
	if isPublic {
		publicStr = "🌍 Public"
	}

	// Use color scheme from commit history
	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Bold(true)
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#707070"))
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#5000ff")).Bold(true)
	//helpTextStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#505050"))

	configContent := titleStyle.Render("Config") + "\n\n" +
		valueStyle.Render(publicStr) + "\n" +
		labelStyle.Render("Default Permission: ") + valueStyle.Render(defaultPerm) // + "\n\n" +
	//helpTextStyle.Render("[t] Toggle Public  [P] Cycle Default Perm")

	configBox := configStyle.Render(configContent)

	// Help section - match commit history help style
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#505050")).
		MarginTop(1)

	help := helpStyle.Render(
		"[n] New User  [d] Delete User  [p] Cycle User Perm [t] Toggle Public  [P] Cycle Default Perm",
	)

	// Layout
	userListView := m.userList.View()

	topSection := lipgloss.JoinHorizontal(
		lipgloss.Top,
		listStyle.Render(userListView),
		configBox,
	)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		topSection,
		help,
	)
}

func (m dashboardModel) renderCreateForm() string {
	formStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#5000ff")).
		Padding(1, 2).
		Width(60)

	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Bold(true)
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#707070"))
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#505050"))
	permStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#5000ff")).Bold(true)

	// Permission is always just displayed, never focused
	permDisplay := permStyle.Render(m.permValue)

	form := titleStyle.Render("Create New User") + "\n\n" +
		labelStyle.Render("Username:") + "\n" +
		m.nameInput.View() + "\n\n" +
		labelStyle.Render("SSH Public Key:") + "\n" +
		m.keyInput.View() + "\n\n" +
		labelStyle.Render("Permission:") + "\n" +
		permDisplay + "\n\n" +
		helpStyle.Render("[↑/↓] Navigate  [p] Cycle Perm  [enter] Create/Next  [esc] Cancel")

	formBox := formStyle.Render(form)

	// Center the form
	if m.width > 0 && m.height > 0 {
		return lipgloss.Place(
			m.width,
			m.height,
			lipgloss.Center,
			lipgloss.Center,
			formBox,
		)
	}
	return formBox
}

func (m dashboardModel) renderDeleteConfirm() string {
	confirmStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#FF1B1C")).
		Padding(1, 2).
		Width(50)

	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF1B1C")).Bold(true)
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#707070"))
	userStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Bold(true)
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#505050"))

	content := titleStyle.Render("Delete User?") + "\n\n" +
		labelStyle.Render("User: ") + userStyle.Render(m.selectedUser) + "\n\n" +
		helpStyle.Render("[d] Delete  [esc] Cancel")

	confirmBox := confirmStyle.Render(content)

	// Center the confirm dialog
	if m.width > 0 && m.height > 0 {
		return lipgloss.Place(
			m.width,
			m.height,
			lipgloss.Center,
			lipgloss.Center,
			confirmBox,
		)
	}
	return confirmBox
}

// Helper functions to modify config/users

func togglePublic() {
	currentPublic := logger.GetConfigPublic()
	newConfig := logger.ConfigData{
		Public:      !currentPublic,
		DefaultPerm: logger.GetConfigDefaultPerm(),
	}
	logger.SetConfig(newConfig)
	if err := logger.WriteJSONFile(logger.Conf, newConfig); err != nil {
		logger.Logger.Error("Failed to write config.json", "error", err)
	} else {
		logger.Logger.Info("config.json updated", "public", newConfig.Public)
	}
}

func cycleDefaultPerm() {
	perms := []string{"none", "read", "write", "admin"}
	current := logger.GetConfigDefaultPerm()

	idx := 0
	for i, p := range perms {
		if p == current {
			idx = (i + 1) % len(perms)
			break
		}
	}

	newConfig := logger.ConfigData{
		Public:      logger.GetConfigPublic(),
		DefaultPerm: perms[idx],
	}
	logger.SetConfig(newConfig)
	if err := logger.WriteJSONFile(logger.Conf, newConfig); err != nil {
		logger.Logger.Error("Failed to write config.json", "error", err)
	} else {
		logger.Logger.Info("config.json updated", "default_perm", newConfig.DefaultPerm)
	}
}

func cyclePermValue(current string, direction int) string {
	perms := []string{"none", "read", "write", "admin"}
	idx := 0

	for i, p := range perms {
		if p == current {
			idx = i
			break
		}
	}

	// Apply direction (1 for forward, -1 for backward)
	idx = (idx + direction + len(perms)) % len(perms)
	return perms[idx]
}

func cycleUserPerm(key string) {
	perms := []string{"none", "read", "write", "admin"}

	users := auth.GetAllUsers()
	user, exists := users[key]
	if !exists {
		return
	}

	idx := 0
	for i, p := range perms {
		if p == user.Perm {
			idx = (i + 1) % len(perms)
			break
		}
	}

	auth.UpdateUserPerm(key, perms[idx])
}

func createUser(key, name, perm string) {
	if perm != "none" && perm != "read" && perm != "write" && perm != "admin" {
		perm = "none"
	}

	auth.AddUser(key, name, perm)
}

func deleteUser(key string) {
	auth.DeleteUser(key)
}
