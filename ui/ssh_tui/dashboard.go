package ssh_tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"strings"

	overlay "github.com/floatpane/bubble-overlay"
	"github.com/nim-sam/gitport/internal/auth"
	"github.com/nim-sam/gitport/internal/git"
	"github.com/nim-sam/gitport/internal/logs"
	"github.com/nim-sam/gitport/pkg/navigator"

	"github.com/nim-sam/gitport/ui"
	"github.com/nim-sam/gitport/ui/ssh_tui/components"
)

// Dashboard states

type dashboardState int

const (
	DefaultState dashboardState = iota
	CreateState
	EditState
	DeleteState
)

// Dashboard Focus states
type dashboardFocus int

const (
	leftPanel dashboardFocus = iota
	rightPanel
	fullPanel
	popUpPanel
)

// Tab states
type tabFocus int

const (
	userManagement tabFocus = iota
	commitHistory
	serverLogs
)

// Dashboard Model

type dashboardModel struct {

	// UserList-related specs
	userList components.UserListModel

	// Dashboard-related specs
	screenWidth  int
	screenHeight int
	innerWidth   int
	innerHeight  int
	state        dashboardState
	focus        dashboardFocus
	tab          tabFocus

	// Panel widths
	leftWidth  int
	rightWidth int

	// Navigator
	navigator navigator.Model

	// Form for new user creation
	userCreationForm components.UserFormModel

	// Form for user information modification
	userModificationForm components.UserFormModel

	// Commit History
	commitHistory components.CommitHistoryModel

	// Server Logs
	serverLogs components.ServerLogsModel

	// Help
	help components.HelpModel
}

// Inititializes and instantiates dashboard mdoel
func NewDashboard(users []*auth.User, commits []*git.Commit, l []*logs.Log) dashboardModel {

	nav := navigator.New()
	nav.Title = "GitPort"
	nav.SetTabs("User Management", "Commit History", "Server Logs")

	u := components.NewUserList(users)

	ucf := components.NewUserCreationFormModel()

	umf := components.NewUserModificationFormModel()

	ch := components.NewCommitHistory(commits)

	sl := components.NewServerLogsModel(l)

	h := components.NewHelp()

	// Return new dashboard model
	return dashboardModel{
		navigator:            nav,
		userList:             u,
		commitHistory:        ch,
		serverLogs:           sl,
		userCreationForm:     ucf,
		userModificationForm: umf,
		help:                 h,
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
		m.innerWidth = int(float32(m.screenWidth) * 0.90)
		m.innerHeight = int(float32(m.screenHeight) * 0.60)

		// Left-Right panel dimensions
		m.leftWidth = int(float32(m.innerWidth) * 0.50)
		m.rightWidth = m.innerWidth - m.leftWidth - 2

		// Navigator inherits width from inner width
		m.navigator.Width = m.innerWidth

		// UserList inherits dimensions from left panel
		m.userList.Height = m.innerHeight
		m.userList.Width = m.leftWidth
		m.userList.ListModel.SetSize(m.userList.Width, m.userList.Height)

		// UserCreationForm inherits dimensions from right panel
		m.userCreationForm.Width = m.rightWidth
		m.userCreationForm.Height = int(float32(m.innerHeight) * 2 / 3)

		// UserModificatoinForm inherits dimensions from right panel
		m.userModificationForm.Width = m.rightWidth
		m.userModificationForm.Height = int(float32(m.innerHeight) * 2 / 3)

		// CommitHistory inherits dimensions from full panel
		m.commitHistory.Height = m.innerHeight
		m.commitHistory.Width = m.leftWidth

		// ServerLogs inherits dimensions from full panel
		m.serverLogs.Width = m.innerWidth
		m.serverLogs.Height = m.innerHeight

		// Help inherits dimenaions from innerWidth
		m.help.Width = m.innerWidth

	case tea.KeyMsg:

		switch msg.String() {
		case "ctrl+q":
			return m, tea.Quit
		// CREATE RESOURCE
		case "ctrl+n":
			switch m.tab {
			// Handling User Creation
			case userManagement:
				switch m.state {
				case CreateState:
					m.state = DefaultState
					m.focus = leftPanel
					m.userCreationForm.ClearInputs()
				case DefaultState:
					m.state = CreateState
				}
			}
		// EDIT RESOURCE
		case "ctrl+e":
			switch m.tab {
			// Handling User Information Modifications
			case userManagement:
				switch m.state {
				case EditState:
					m.state = DefaultState
					m.focus = leftPanel
					m.userModificationForm.ClearInputs()
				case DefaultState:
					m.state = EditState

					// Fetch currently selected user and populate data in pop-up
					currentUserIndex := m.userList.ListModel.Index()
					currentUser := m.userList.UserData[currentUserIndex]
					m.userModificationForm.Inputs[0].SetValue(currentUser.Name())
					m.userModificationForm.Inputs[1].SetValue(currentUser.SshKey())
					m.userModificationForm.Selector.SetValue(int(currentUser.AccessLevel()))
				}
			}

		case "esc":
			m.state = DefaultState
			m.userCreationForm.ClearInputs()
			m.focus = leftPanel
		// Swtiching focus between left and right panels
		case "enter":
			if m.focus == leftPanel && m.state == DefaultState {
				m.focus = rightPanel
			} else if m.focus == rightPanel && m.state == DefaultState {
				m.focus = leftPanel
			}
		}
	}

	// Handling tab selection

	previousTab := m.tab

	m.navigator, _ = m.navigator.Update(msg)
	m.tab = tabFocus(m.navigator.Index())

	// Reset Focus when switching tabs

	if m.tab != previousTab {
		switch m.tab {
		case userManagement, commitHistory:
			m.focus = leftPanel
		case serverLogs:
			m.focus = fullPanel
		}
	}

	// Passing messages to dashboard sub-models

	switch m.tab {
	case userManagement:
		switch m.state {
		// Broswing users
		case DefaultState:
			if m.focus == leftPanel {
				m.userList.ListModel, _ = m.userList.ListModel.Update(msg)
			}
		// Creating user
		case CreateState:
			m.focus = popUpPanel
			m.userCreationForm.UpdateInputs(msg)
			m.userCreationForm.Selector.UpdateSelection(msg)
		// Editing user
		case EditState:
			m.focus = popUpPanel
			m.userModificationForm.UpdateInputs(msg)
			m.userModificationForm.Selector.UpdateSelection(msg)
		}
	case commitHistory:
		// m.focus = leftPanel
		m.commitHistory, _ = m.commitHistory.Update(msg)
	case serverLogs:
		m.focus = fullPanel
		m.serverLogs, _ = m.serverLogs.Update(msg)
	}

	m.help, _ = m.help.Update(msg)
	m.userCreationForm, _ = m.userCreationForm.Update(msg)
	m.userModificationForm, _ = m.userModificationForm.Update(msg)

	return m, nil
}

func (m dashboardModel) View() string {

	var leftPanelContent string
	var rightPanelContent string
	var fullPanelContent string

	// Tab View Content Management

	switch m.tab {
	// User Management Tab
	case userManagement:
		switch m.state {
		case DefaultState:
			leftPanelContent = m.userList.View()
			rightPanelContent = ""
		case CreateState:
			leftPanelContent = m.userList.View()
			rightPanelContent = ""
		case EditState:
			leftPanelContent = m.userList.View()
			rightPanelContent = ""
		}
	// Commit History Tab
	case commitHistory:
		switch m.state {
		default:
			leftPanelContent = m.commitHistory.View()
			rightPanelContent = ""
		}
	// Server Logs Tab
	case serverLogs:
		switch m.state {
		default:
			fullPanelContent = m.serverLogs.View()
		}
	}

	// Navigator styling

	m.navigator.Styles.Title = m.navigator.Styles.Title.Foreground(ui.Palette.Accent)
	m.navigator.Styles.TabBlurred = m.navigator.Styles.TabBlurred.Foreground(ui.Palette.FgMuted).Faint(true)
	m.navigator.Styles.TabFocused = m.navigator.Styles.TabFocused.Foreground(ui.Palette.Fg).BorderForeground(ui.Palette.Fg)
	m.navigator.Styles.Separator = m.navigator.Styles.Separator.Foreground(ui.Palette.FgMuted).Bold(true)
	navigatorContainer := lipgloss.NewStyle().Padding(0, 1)

	// Panel Styling

	var fullPanelStyle lipgloss.Style
	var leftPanelStyle lipgloss.Style
	var rightPanelStyle lipgloss.Style

	panelStyleBlurred := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ui.Palette.Border).
		Height(m.innerHeight).
		AlignHorizontal(lipgloss.Left)

	panelStyleFocused := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ui.Palette.SelectedBorder).
		Height(m.innerHeight).
		AlignHorizontal(lipgloss.Left)

	// Panel styling per tab and panel focus

	switch m.tab {
	case userManagement, commitHistory:
		switch m.focus {
		case leftPanel:
			leftPanelStyle = panelStyleFocused.Width(m.leftWidth)
			rightPanelStyle = panelStyleBlurred.Width(m.rightWidth)
		case rightPanel:
			leftPanelStyle = panelStyleBlurred.Width(m.leftWidth)
			rightPanelStyle = panelStyleFocused.Width(m.rightWidth)
		}
	case serverLogs:
		switch m.focus {
		case fullPanel:
			fullPanelStyle = panelStyleFocused.
				Width(m.innerWidth).
				AlignVertical(lipgloss.Center).
				AlignHorizontal(lipgloss.Center).
				Foreground(ui.Palette.FgMuted)
		default:
			fullPanelStyle = panelStyleBlurred.
				Width(m.innerWidth).
				AlignVertical(lipgloss.Center).
				Foreground(ui.Palette.FgMuted)
		}
	}

	// Pop-up window styling
	popUpStyle := lipgloss.NewStyle().
		Width(m.rightWidth).
		Height(int(float32(m.innerHeight) * 2.0 / 3.0)).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ui.Palette.Border)

	if m.focus == popUpPanel {
		popUpStyle = popUpStyle.BorderForeground(ui.Palette.SelectedBorder)
		leftPanelStyle = panelStyleBlurred.Width(m.leftWidth)
		rightPanelStyle = panelStyleBlurred.Width(m.rightWidth)
	}

	// HelpBar Styling
	helpBarStyle := lipgloss.NewStyle().
		Width(m.innerWidth).
		Padding(1, 0, 0, 0)

	// Terminal window-sized application container, that will be containing
	// the left and right panels, centered vertically and horizontally
	dashBoardStyle := lipgloss.NewStyle().
		Width(m.screenWidth).
		Height(m.screenHeight).
		Align(lipgloss.Center, lipgloss.Center)

	// Aggregating panels and UI elements to form dashboard view

	var screenSetup string

	fullScreenSetup := fullPanelStyle.Render(fullPanelContent)

	splitScreenSetup := lipgloss.JoinHorizontal(
		lipgloss.Left,
		leftPanelStyle.Render(leftPanelContent),
		rightPanelStyle.Render(rightPanelContent),
	)

	// Setup configuarion with respect to tab focused

	switch m.tab {
	case userManagement:
		screenSetup = splitScreenSetup
	case commitHistory:
		screenSetup = splitScreenSetup
	case serverLogs:
		screenSetup = fullScreenSetup
	}

	output := dashBoardStyle.Render(
		lipgloss.JoinVertical(lipgloss.Top,
			navigatorContainer.Render(m.navigator.View()),
			screenSetup,
			helpBarStyle.Render(m.help.View()),
		))

	// Injecting Pop-up window
	form := m.userCreationForm
	switch m.state {
	case CreateState:
		form = m.userCreationForm
	case EditState:
		form = m.userModificationForm
	}
	toInject := strings.Split(popUpStyle.Render(form.View()), "\n")

	x_coord := int((1.0 / 2.0) * float32(m.screenWidth-form.Width-2))
	y_coord := int((1.0 / 2.0) * float32(m.screenHeight-form.Height-2))

	if m.state == CreateState || m.state == EditState {
		output = overlay.Block(output, toInject, y_coord, x_coord)
	}

	return output
}
