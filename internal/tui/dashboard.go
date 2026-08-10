package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	overlay "github.com/floatpane/bubble-overlay"
	"github.com/nim-sam/gitport/internal/auth"
	"strings"
)

// Dashboard states

type dashboardState int

const (
	DefaultState dashboardState = iota
	UserCreateState
	UserDeleteState
)

// Dashboard Focus states
type dashhboardFocus int

const (
	leftPanel dashhboardFocus = iota
	rightPanel
	popUpPanel
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
	focus        dashhboardFocus

	// Panel widths
	leftWidth  int
	rightWidth int

	// Form for new user creation
	userCreationForm UserFormModel

	// Help
	help HelpModel
}

// Inititializes and instantiates dashboard mdoel
func NewDashboard(users []*auth.User) dashboardModel {

	u := newUserList(users)

	ucf := newUserFormModel()

	h := newHelp()

	// Return new dashboard model
	return dashboardModel{
		userList:         u,
		userCreationForm: ucf,
		help:             h,
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
		m.innerWidth = int(float32(m.screenWidth) * 0.85)
		m.innerHeight = int(float32(m.screenHeight) * 0.85)

		// Left-Right panel dimensions
		m.leftWidth = int(float32(m.innerWidth)*0.50) - 1
		m.rightWidth = m.innerWidth - m.leftWidth - 1

		// UserList inherits diemensions from left panel
		m.userList.height = m.innerHeight
		m.userList.width = m.leftWidth
		m.userList.listModel.SetSize(m.userList.width, m.userList.height)

		// UserCreationForm inheritds dimensions from right panel
		m.userCreationForm.width = m.rightWidth
		m.userCreationForm.heigt = int(float32(m.innerHeight) * 2 / 3)

		for i := range m.userCreationForm.inputs {
			prompt := m.userCreationForm.inputs[i].Prompt
			m.userCreationForm.inputs[i].Width = m.userCreationForm.width - 6 - len(prompt) - 1
		}

		// Help inherits dimenaions from innerWidth
		m.help.width = m.innerWidth

	case tea.KeyMsg:

		switch msg.String() {
		case "ctrl+q":
			return m, tea.Quit
		case "ctrl+n":
			if m.state == UserCreateState {
				m.state = DefaultState
				m.focus = leftPanel
				m.userCreationForm.ClearInputs()
			} else {
				m.state = UserCreateState
			}
		case "esc":
			m.state = DefaultState
			m.userCreationForm.ClearInputs()
			m.focus = leftPanel
		case "tab":
			if m.focus == leftPanel && m.state == DefaultState {
				m.focus = rightPanel
			} else if m.focus == rightPanel && m.state == DefaultState {
				m.focus = leftPanel
			}
		}

	}

	// var cmd tea.Cmd
	if m.focus == leftPanel && m.state == DefaultState {
		m.userList.listModel, _ = m.userList.listModel.Update(msg)
	}

	if m.state == UserCreateState {
		m.focus = popUpPanel
		m.userCreationForm.UpdateInputs(msg)
		m.userCreationForm.selector.UpdateSelection(msg)
	}

	m.help, _ = m.help.Update(msg)

	return m, nil
}

func (m dashboardModel) View() string {

	leftPanelModel := m.userList

	var rightPanelModel tea.Model

	switch m.state {
	case DefaultState:
		rightPanelModel = m.userList
	case UserCreateState:
		rightPanelModel = m.userList
	}

	// Left panel styling
	leftPanelStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(palette.border).
		Height(m.innerHeight).
		Width(m.leftWidth).
		AlignHorizontal(lipgloss.Center).
		Padding(0, 1)

	if m.focus == leftPanel {
		leftPanelStyle = leftPanelStyle.BorderForeground(palette.selectedBorder)
	}

	_ = leftPanelStyle

	// Right Panel styling
	rightPanelStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(palette.border).
		Height(m.innerHeight).
		Width(m.rightWidth)

	if m.focus == rightPanel {
		rightPanelStyle = rightPanelStyle.BorderForeground(palette.selectedBorder)
	}

	// Pop-up window styling
	popUpStyle := lipgloss.NewStyle().
		Width(m.rightWidth).
		Height(int(float32(m.innerHeight) * 2 / 3)).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(palette.border)

	if m.focus == popUpPanel {
		popUpStyle = popUpStyle.BorderForeground(palette.selectedBorder)
	}

	// Terminal window-sized application container, that will be containing
	// the left and right panels, centered vertically and horizontally
	dashBoardStyle := lipgloss.NewStyle().
		Width(m.screenWidth).
		Height(m.screenHeight).
		Align(lipgloss.Center, lipgloss.Center)

	// Aggregating the left and right panels inside dashBoardStyle
	output := dashBoardStyle.Render(

		lipgloss.JoinVertical(lipgloss.Top,

			lipgloss.JoinHorizontal(
				lipgloss.Left,
				leftPanelStyle.Render(leftPanelModel.View()),
				rightPanelStyle.Render(rightPanelModel.View()),
			),
			m.help.View(),
		))

	// Injecting Pop-up window
	toInject := strings.Split(popUpStyle.Render(m.userCreationForm.View()), "\n")

	x_coord := int((1.0 / 2.0) * float32(m.screenWidth-m.userCreationForm.width-2))
	y_coord := int((1.0 / 2.0) * float32(m.screenHeight-m.userCreationForm.heigt-2))

	if m.state == UserCreateState {
		output = overlay.Block(output, toInject, y_coord, x_coord)
	}

	return output
}
