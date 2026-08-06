package tui

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/nim-sam/gitport/internal/auth"
)

// Dashboard states

type dashboardState int

const (
	DefaultState dashboardState = iota
	UserCreateState
	UserDeleteState
)

// Dashboard Model

type dashboardModel struct {

	// UserList-related specs
	userList UserListModel

	// Dashboard-related specs
	screenWidth  int
	screenHeight int
	innerWidth   int
	innerHeight  int
	state        dashboardState

	// Panel widths
	leftWidth  int
	rightWidth int

	// Form inputs for new user creation
	nameInput       textinput.Model
	sshKeyInput     textinput.Model
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

		// Terminal window dimensions
		m.screenWidth = msg.Width - 2
		m.screenHeight = msg.Height - 2

		// Application interface dimensions
		m.innerWidth = int(float32(m.screenWidth) * 0.75)
		m.innerHeight = int(float32(m.screenHeight) * 0.75)

		// Left-Right panel dimensions
		m.leftWidth = int(float32(m.innerWidth)*0.50) - 1
		m.rightWidth = m.innerWidth - m.leftWidth - 1

		// UserList inherits diemensions from left panel
		m.userList.height = m.innerHeight
		m.userList.width = m.leftWidth
		m.userList.listModel.SetSize(m.userList.width, m.userList.height)

	case tea.KeyMsg:

		switch msg.String() {
		case "q":
			return m, tea.Quit
		case "n":
			m.state = UserCreateState
		case "esc":
			m.state = DefaultState
		}

	}

	// var cmd tea.Cmd
	m.userList.listModel, _ = m.userList.listModel.Update(msg)

	return m, nil
}

func (m dashboardModel) View() string {

	// Left panel styling
	leftPanelStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(palette.border).
		Height(m.innerHeight).
		Width(m.leftWidth).
		AlignHorizontal(lipgloss.Center).
		Padding(0, 1)

	_ = leftPanelStyle

	// Right Panel styling
	rightPanelStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(palette.border).
		Height(m.innerHeight).
		Width(m.rightWidth)

	// Terminal window-sized application container, that will be containing
	// the left and right panels, centered vertically and horizontally
	dashBoardStyle := lipgloss.NewStyle().
		Width(m.screenWidth).
		Height(m.screenHeight).
		Align(lipgloss.Center, lipgloss.Center)

	// Aggregating the left and right panels inside dashBoardStyle
	output := dashBoardStyle.Render(lipgloss.JoinHorizontal(
		lipgloss.Left,
		leftPanelStyle.Render(m.userList.View()),
		rightPanelStyle.String(),
	))

	return output
}
