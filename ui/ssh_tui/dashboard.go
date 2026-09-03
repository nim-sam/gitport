package ssh_tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"strings"

	overlay "github.com/floatpane/bubble-overlay"
	"github.com/nim-sam/gitport/internal/repository"
	"github.com/nim-sam/gitport/internal/server"
	"github.com/nim-sam/gitport/pkg/dialog"
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
	FilterState
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

	// Dialog windows
	confirmUserDeletionDialog dialog.Model

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
func NewDashboard(server server.Server) dashboardModel {

	// Loading records from database
	users := repository.UserRepo.Load()
	commits := server.LoadCommits(30)
	logs := repository.LogRepo.Load()

	nav := navigator.New()
	nav.Title = "GitPort"
	nav.SetTabs("User Management", "Commit History", "Server Logs")

	dia := dialog.New()

	u := components.NewUserList(users)

	ucf := components.NewUserCreationFormModel()

	umf := components.NewUserModificationFormModel()

	ch := components.NewCommitHistory(commits)

	sl := components.NewServerLogsModel(logs)

	h := components.NewHelp()

	// Return new dashboard model
	return dashboardModel{
		navigator:                 nav,
		userList:                  u,
		commitHistory:             ch,
		serverLogs:                sl,
		userCreationForm:          ucf,
		userModificationForm:      umf,
		confirmUserDeletionDialog: dia,
		help:                      h,
	}
}

// dashboardModel's Model implementation

func (m dashboardModel) Init() tea.Cmd {
	return nil
}

func (m dashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	var cmd tea.Cmd

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

		// Dialog inherits dimensions from right panel
		m.confirmUserDeletionDialog.Width = m.rightWidth
		m.confirmUserDeletionDialog.Height = int(float32(m.innerHeight) * 2 / 3)

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
					m.userModificationForm.ResetSelection()

					// Fetch currently selected user and populate data in pop-up
					currentUser := m.userList.GetSelectedUser()
					m.userModificationForm.SetUserFields(currentUser)
				}
			}
		// DELETE RESOURCE
		case "ctrl+d":
			switch m.tab {
			case userManagement:
				switch m.state {
				case DefaultState:

					m.state = DeleteState
					user := m.userList.GetSelectedUser()

					// Set dialog data catered for user deletion
					m.confirmUserDeletionDialog.ResetSelection()
					m.confirmUserDeletionDialog.Title = "Delete User"
					m.confirmUserDeletionDialog.Prompt = fmt.Sprintf("Are you sure you want to delete %s?", user.Name)
					m.confirmUserDeletionDialog.ConfirmMsg = "[ Confirm ]"
					m.confirmUserDeletionDialog.CancelMsg = "[ Cancel ]"

				}
			}
		case "esc":
			m.state = DefaultState
			m.userCreationForm.ClearInputs()
			m.focus = leftPanel
		// Switching focus between left and right panels
		case "enter":
			switch m.state {
			case DefaultState:
				switch m.focus {
				case leftPanel:
					m.focus = rightPanel
				case rightPanel:
					m.focus = leftPanel
				}
			}
		case "tab":
			switch m.state {
			// Swtich tabs when in default state
			case DefaultState:
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
			}
		}
	// Handles user deletion dialog
	case dialog.SelectionMsg:
		if msg.Confirm && m.state == DeleteState {
			// Fetches selected user, deletes user from data base,
			// updates user list model
			user := m.userList.GetSelectedUser()
			repository.UserRepo.Delete(user.ID)
			usersUpdated := repository.UserRepo.Load()
			m.userList.UpdateUsers(usersUpdated)

			// Return to default state
			m.state = DefaultState
			m.focus = leftPanel
		} else {
			m.state = DefaultState
			m.focus = leftPanel
		}
	// Handles user creation/modification forms
	case components.UserFormConfirmationMsg:
		switch m.state {
		case CreateState:
			// Save user and reload users from database
			newUser := msg.GetUser()
			repository.UserRepo.Save(newUser)
			m.userList.UpdateUsers(repository.UserRepo.Load())

			// Clear inupts, reset selection
			m.userCreationForm.ClearInputs()
			m.userCreationForm.ResetSelection()

			// Return to default state
			m.state = DefaultState
			m.focus = leftPanel
		case EditState:
			// Updates fields of currently selected user and reload
			// users from database
			currentUser := m.userList.GetSelectedUser()
			userID := currentUser.ID
			updatedUser := msg.GetUser()
			updatedUser.ID = userID
			repository.UserRepo.Save(updatedUser)
			m.userList.UpdateUsers(repository.UserRepo.Load())

			// Return to default state
			m.state = DefaultState
			m.focus = leftPanel
		}
	}

	// Passing messages to dashboard sub-models

	switch m.tab {
	case userManagement:
		switch m.state {
		// Broswing users
		case DefaultState, FilterState:
			if m.focus == leftPanel {
				m.userList, _ = m.userList.Update(msg)
			}
		// Creating user
		case CreateState:
			m.focus = popUpPanel
			m.userCreationForm, cmd = m.userCreationForm.Update(msg)
		// Editing user
		case EditState:
			m.focus = popUpPanel
			m.userModificationForm, cmd = m.userModificationForm.Update(msg)
		// Deleting user dialog
		case DeleteState:
			m.focus = popUpPanel
			m.confirmUserDeletionDialog, cmd = m.confirmUserDeletionDialog.Update(msg)

		}
	case commitHistory:
		// m.focus = leftPanel
		if m.focus == leftPanel {
			m.commitHistory, _ = m.commitHistory.Update(msg)
		}
	case serverLogs:
		m.focus = fullPanel
		m.serverLogs, _ = m.serverLogs.Update(msg)
	}

	// Handling navigation during filter states from list models
	if (m.userList.ListModel.FilterState() == list.Filtering ||
		m.commitHistory.ListModel.FilterState() == list.Filtering) &&
		m.state == DefaultState {
		m.state = FilterState
	} else if m.userList.ListModel.FilterState() == list.FilterApplied ||
		m.commitHistory.ListModel.FilterState() == list.FilterApplied {
		m.state = DefaultState
	}

	m.help, _ = m.help.Update(msg)

	return m, cmd
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
		default:
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

	// Confirm User Deletion dialog window styling
	m.confirmUserDeletionDialog.Style.Title = lipgloss.NewStyle().
		Foreground(ui.Palette.FgHighlight).
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(ui.Palette.Border).
		Bold(true)
	m.confirmUserDeletionDialog.Style.Prompt = lipgloss.NewStyle().
		Foreground(ui.Palette.Fg)
	m.confirmUserDeletionDialog.Style.ConfirmMsgNormal = lipgloss.NewStyle().
		Foreground(ui.Palette.FgMuted)
	m.confirmUserDeletionDialog.Style.ConfirmMsgSelected = lipgloss.NewStyle().
		Foreground(ui.Palette.Accent)
	m.confirmUserDeletionDialog.Style.CancelMsgNormal = lipgloss.NewStyle().
		Foreground(ui.Palette.FgMuted)
	m.confirmUserDeletionDialog.Style.CancelMsgSelected = lipgloss.NewStyle().
		Foreground(ui.Palette.Accent)

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
	var popUpContent string
	var popUpWidth int
	var popUpHeight int

	switch m.state {
	case CreateState:
		popUpContent = m.userCreationForm.View()
		popUpWidth = m.userCreationForm.Width
		popUpHeight = m.userCreationForm.Height
	case EditState:
		popUpContent = m.userModificationForm.View()
		popUpWidth = m.userModificationForm.Width
		popUpHeight = m.userModificationForm.Height
	case DeleteState:
		popUpContent = m.confirmUserDeletionDialog.View()
		popUpWidth = m.confirmUserDeletionDialog.Width
		popUpHeight = m.confirmUserDeletionDialog.Height
	}
	toInject := strings.Split(popUpStyle.Render(popUpContent), "\n")

	x_coord := int((1.0 / 2.0) * float32(m.screenWidth-popUpWidth-2))
	y_coord := int((1.0 / 2.0) * float32(m.screenHeight-popUpHeight-2))

	if m.state == CreateState || m.state == EditState || m.state == DeleteState {
		output = overlay.Block(output, toInject, y_coord, x_coord)
	}

	return output
}
