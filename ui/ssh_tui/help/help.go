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
	Width   int
	Height  int

	// admin gates the full help view. Non-admins only get the short help bar:
	// the user-management bindings and the "?" toggle are removed entirely.
	admin bool
}

// New builds the help model. When admin is false the user-management
// keybindings and the help toggle are disabled so they never appear; the
// package-level keys are copied first so disabling never leaks across sessions.
func New(admin bool) Model {

	// TUI opens user management tab on startup, so we set user management keys
	// as defualt keybinds for the help bar
	k := userManagementKeys

	if !admin {
		k.NewUser.SetEnabled(false)
		k.EditUser.SetEnabled(false)
		k.DeleteUser.SetEnabled(false)
		k.Help.SetEnabled(false)
	}

	return Model{
		keys:   k,
		help:   help.New(),
		Height: 2,
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
			if m.admin {
				m.help.ShowAll = !m.help.ShowAll
			}
		}
	// Receives current tab from NavigatorMsg directly
	case navigator.NavigatorMsg:
		switch msg.GetCurrentTab() {
		case 0:
			m.keys = userManagementKeys
		case 1:
			m.keys = commitHistoryKeys
		case 2:
			m.keys = serverLogsKeys
		}
	}

	return m, nil
}

func (m Model) View() string {

	// Non-admins have no full help view, so never render it even if the help
	// model was left in the expanded state.
	if !m.admin {
		m.help.ShowAll = false
	}

	helpContainer := lipgloss.NewStyle().Width(m.Width).Height(m.Height).AlignHorizontal(lipgloss.Center)

	return helpContainer.Render(m.help.View(m.keys))
}
