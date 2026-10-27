package help

import (
	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nim-sam/gitport/pkg/navigator"
)

type Model struct {
	keys    DashboardKeys
	help    help.Model
	lastKey string
	width   int
	height  int

	// admin gates the full help view. Non-admins only get the short help bar:
	// the user-management bindings and the "?" toggle are removed entirely.
	admin bool

	// internal value storage for current tab
	currentTab int
}

// Dimension setters
func (m *Model) Width(width int)   { m.width = width }
func (m *Model) Height(height int) { m.height = height }

// New builds the help model. When admin is false the user-management
// keybindings and the help toggle are disabled so they never appear; the
// package-level keys are copied first so disabling never leaks across sessions.
func New(admin bool) Model {

	// The TUI opens the server overview tab on startup, so its keys are the
	// default shown in the help bar.

	var k DashboardKeys
	if admin {
		k = serverOverviewKeysAdmin
	} else {
		k = serverOverviewKeysNonAdmin
	}

	return Model{
		keys:   k,
		help:   help.New(),
		height: 2,
		admin:  admin,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up":
			m.lastKey = "↑"
		case "down":
			m.lastKey = "↓"
		case "tab":
			m.lastKey = "tab"
		case "ctrl+n":
			m.lastKey = "ctrl+n"
		case "ctrl+e":
			m.lastKey = "ctrl+e"
		case "ctrl+d":
			m.lastKey = "ctrl+d"
		case "ctrl+q":
			m.lastKey = "ctrl+q"
		case "enter":
			m.lastKey = "enter"
		case "/":
			m.lastKey = "/"
		case "?":
			// if m.admin && m.currentTab == 0 {
			// 	m.help.ShowAll = !m.help.ShowAll
			// }
			switch m.currentTab {
			case 0:
				if m.admin {
					m.help.ShowAll = !m.help.ShowAll
				}
			case 1:
				if m.admin {
					m.help.ShowAll = !m.help.ShowAll
				}
			case 2:
				m.help.ShowAll = !m.help.ShowAll
			}
		}
	// Receives current tab from NavigatorMsg directly
	case navigator.NavigatorMsg:
		// Resent show all toggle
		m.help.ShowAll = false

		// Switch to appropriate key bind set
		m.currentTab = msg.GetCurrentTab()
		switch m.currentTab {
		case 0:
			if m.admin {
				m.keys = serverOverviewKeysAdmin
			} else {
				m.keys = serverOverviewKeysNonAdmin
			}
		case 1:
			if m.admin {
				m.keys = userManagementKeysAdmin
			} else {
				m.keys = userManagementKeysNonAdmin
			}
		case 2:
			m.keys = commitHistoryKeys
		case 3:
			m.keys = serverLogsKeys
		}
	}

	return m, nil
}

func (m Model) View() string {

	helpContainer := lipgloss.NewStyle().Width(m.width).Height(m.height).AlignHorizontal(lipgloss.Center)

	return helpContainer.Render(m.help.View(m.keys))
}
