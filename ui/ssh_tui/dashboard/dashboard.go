package dashboard

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"strings"

	overlay "github.com/floatpane/bubble-overlay"
	"github.com/nim-sam/gitport/internal/model"
	"github.com/nim-sam/gitport/internal/repository"
	"github.com/nim-sam/gitport/pkg/dialog"
	"github.com/nim-sam/gitport/pkg/navigator"

	"github.com/nim-sam/gitport/ui"
	commithistory "github.com/nim-sam/gitport/ui/ssh_tui/commit_history"
	commitviewport "github.com/nim-sam/gitport/ui/ssh_tui/commit_viewport"
	"github.com/nim-sam/gitport/ui/ssh_tui/form"
	"github.com/nim-sam/gitport/ui/ssh_tui/help"
	serverlogs "github.com/nim-sam/gitport/ui/ssh_tui/server_logs"
	userlist "github.com/nim-sam/gitport/ui/ssh_tui/user_list"
)

/*
 * The Dashboard model is the ssh-accessible Terminal User Interface (TUI) that handles
 * GitPort server management. The TUI provides admin users the ability to manage users
 * (creating, editing, deleting), view most recent commits done on the repository along
 * their respective commit diffs, and a table of timestamped server logs.
 */

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

type Model struct {

	// UserList-related specs
	userList userlist.Model

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

	// Git diff commitViewport
	commitViewport commitviewport.Model

	// Navigator
	navigator navigator.Model

	// Dialog windows
	confirmUserDeletionDialog dialog.Model

	// Form for new user creation
	userCreationForm form.Model

	// Form for user information modification
	userModificationForm form.Model

	// Commit History
	commitHistory commithistory.Model

	// Server Logs
	serverLogs serverlogs.Model

	// Help
	help help.Model
}

// Fallback terminal geometry, used when a session reports no usable PTY
// dimensions. Without it, sub-model widths would derive from zero and the
// dashboard would render collapsed.
const (
	fallbackWidth  = 80
	fallbackHeight = 24
)

// Inititializes and instantiates dashboard mdoel
//
// width and height are the dimensions of the terminal the dashboard is being
// rendered into. Over ssh they come from the session's PTY, since Bubble Tea
// cannot detect the size of a non-local terminal on startup.
//
// diff resolves a commit hash to its unified diff. It is supplied by the caller
// rather than imported, because the ssh server that assembles this dashboard
// lives in internal/server and must not be depended upon from here.
func New(users []model.User, commits []model.Commit, logs []model.Log, width, height int, diff commitviewport.DiffSource) Model {

	nav := navigator.New()
	nav.Title = "GitPort"
	nav.SetTabs("User Management", "Commit History", "Server Logs")

	dia := dialog.New()

	u := userlist.New(users)

	ucf := form.NewUserCreationFormModel()

	umf := form.NewUserModificationFormModel()

	ch := commithistory.New(commits)

	sl := serverlogs.New(logs)

	h := help.New()

	v := commitviewport.New(make(map[string]string), diff)

	// Return new dashboard model
	m := Model{
		commitViewport:            v,
		navigator:                 nav,
		userList:                  u,
		commitHistory:             ch,
		serverLogs:                sl,
		userCreationForm:          ucf,
		userModificationForm:      umf,
		confirmUserDeletionDialog: dia,
		help:                      h,
	}

	// Lay the dashboard out immediately so the very first frame is sized
	// correctly instead of waiting for a tea.WindowSizeMsg that only arrives
	// once the terminal is resized.
	m.resize(width, height)

	return m
}

// resize propagates a terminal resize down to every sub-model, deriving the
// dashboard, inner interface and panel dimensions from the given screen size.
func (m *Model) resize(width, height int) {
	if width <= 0 {
		width = fallbackWidth
	}
	if height <= 0 {
		height = fallbackHeight
	}

	// Terminal window dimensions
	m.screenWidth = width - 2
	m.screenHeight = height - 2

	// Application interface dimensions
	m.innerWidth = int(float32(m.screenWidth) * 0.90)
	m.innerHeight = int(float32(m.screenHeight) * 0.60)

	// Left-Right panel dimensions
	m.leftWidth = int(float32(m.innerWidth) * 0.50)
	m.rightWidth = m.innerWidth - m.leftWidth - 2

	// commitViewport inherits dimensions from right panel
	m.commitViewport.Height(m.innerHeight)
	m.commitViewport.Width(m.rightWidth)

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
}

// Model's Model implementation

func (m Model) Init() tea.Cmd {
	return m.commitHistory.Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	var cmds tea.BatchMsg

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:

		// Propagate the new terminal dimensions to every sub-model
		m.resize(msg.Width, msg.Height)

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
	case form.UserFormConfirmationMsg:
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
	var SubmodelUpdateCmd tea.Cmd
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
			m.userCreationForm, SubmodelUpdateCmd = m.userCreationForm.Update(msg)
		// Editing user
		case EditState:
			m.focus = popUpPanel
			m.userModificationForm, SubmodelUpdateCmd = m.userModificationForm.Update(msg)
		// Deleting user dialog
		case DeleteState:
			m.focus = popUpPanel
			m.confirmUserDeletionDialog, SubmodelUpdateCmd = m.confirmUserDeletionDialog.Update(msg)
		}
		cmds = append(cmds, SubmodelUpdateCmd)

	case commitHistory:

		if m.focus == leftPanel {
			m.commitHistory.SetFocused(true)
			m.commitViewport.SetFocused(false)
		} else {
			m.commitHistory.SetFocused(false)
			m.commitViewport.SetFocused(true)
		}

		// Update commit history list model
		var commitHistoryMessage tea.Cmd
		m.commitHistory, commitHistoryMessage = m.commitHistory.Update(msg)
		cmds = append(cmds, commitHistoryMessage)

		// Update commit viewport
		var commitViewportMessage tea.Cmd
		m.commitViewport, commitViewportMessage = m.commitViewport.Update(msg)
		cmds = append(cmds, commitViewportMessage)

	case serverLogs:
		m.focus = fullPanel
		m.serverLogs, SubmodelUpdateCmd = m.serverLogs.Update(msg)
		cmds = append(cmds, SubmodelUpdateCmd)
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

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {

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
			rightPanelContent = m.commitViewport.View()
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
