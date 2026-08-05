package tui

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/nim-sam/gitport/pkg/auth"
)

// Dashboard states

type dashboardState int

const (
	defaultState dashboardState = iota
	createState
	deleteState
)

// Dashboard Model

type dashboardModel struct {
	userList     UserListModel
	selectedUser int
	selectedKey  string
	width        int
	height       int

	state dashboardState

	// Form inputs for new user creation
	nameInput   textinput.Model
	sshKeyInput textinput.Model
	// permLevel       string
	userListFocused bool
}

// Inititializes and instantiates dashboard mdoel
func NewDashboard(users []*auth.User) dashboardModel {

	u := newUserList(users)

	// Creating form inputs
	nameInput := textinput.New()
	nameInput.Placeholder = "Enter username..."
	nameInput.CharLimit = 64

	sshKeyInput := textinput.New()
	sshKeyInput.Placeholder = "Paste public SSH key..."
	sshKeyInput.CharLimit = 512

	// Return new dashboard model
	return dashboardModel{
		userList:    u,
		nameInput:   nameInput,
		sshKeyInput: sshKeyInput,
		// permLevel:       "none",
		userListFocused: true,
	}
}

// dashboardModel's Model implementation

func (m dashboardModel) Init() tea.Cmd {
	return nil
}

func (m dashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width - 2
		m.height = msg.Height - 2

		m.userList.height = m.height
		m.userList.width = int(float32(m.width) * 0.40)

		m.userList.listModel.SetSize(m.userList.width, m.userList.height)

	case tea.KeyMsg:

		switch msg.String() {
		case "q":
			tea.Quit()
		}

	}

	// var cmd tea.Cmd
	m.userList.listModel, _ = m.userList.listModel.Update(msg)

	return m, nil
}

func (m dashboardModel) View() string {

	dashBoardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(palette.border).
		Height(m.height).
		Width(m.width)

	// listWidth := int(float64(m.width) * 0.9)

	output := dashBoardStyle.Render(lipgloss.JoinHorizontal(
		lipgloss.Left,
		m.userList.View(),
	))

	return output
}
