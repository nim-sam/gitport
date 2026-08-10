package tui

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type DashboardKeys struct {
	Up         key.Binding
	Down       key.Binding
	Tab        key.Binding
	NewUser    key.Binding
	EditUser   key.Binding
	DeleteUser key.Binding
	Quit       key.Binding
	Help       key.Binding
}

func (d DashboardKeys) ShortHelp() []key.Binding {
	return []key.Binding{d.Up, d.Down, d.Tab, d.Quit, d.Help}
}

func (d DashboardKeys) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{d.Up, d.Down},
		{d.Tab, d.Quit},
		{d.NewUser, d.EditUser},
		{d.DeleteUser},
		{d.Quit, d.Help},
	}
}

var keys = DashboardKeys{
	Up: key.NewBinding(
		key.WithKeys("up"),
		key.WithHelp("↑", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down"),
		key.WithHelp("↓", "move down"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "switch panel"),
	),
	NewUser: key.NewBinding(
		key.WithKeys("ctrl+n"),
		key.WithHelp("ctrl+n", "create user"),
	),
	EditUser: key.NewBinding(
		key.WithKeys("ctrl+e"),
		key.WithHelp("ctrl+e", "edit user"),
	),
	DeleteUser: key.NewBinding(
		key.WithKeys("ctrl+d"),
		key.WithHelp("ctrl+d", "delete user"),
	),
	Quit: key.NewBinding(
		key.WithKeys("ctrl+q"),
		key.WithHelp("ctrl+q", "quit"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
}

type HelpModel struct {
	keys    DashboardKeys
	help    help.Model
	lastKey string
	width   int
}

func newHelp() HelpModel {
	return HelpModel{
		keys: keys,
		help: help.New(),
	}
}

func (m HelpModel) Init() tea.Cmd {
	return nil
}

func (m HelpModel) Update(msg tea.Msg) (HelpModel, tea.Cmd) {

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
		case "?":
			m.help.ShowAll = !m.help.ShowAll
		}
	}

	return m, nil
}

func (m HelpModel) View() string {

	helpContainer := lipgloss.NewStyle().Width(m.width).AlignHorizontal(lipgloss.Center)

	return helpContainer.Render(m.help.View(m.keys))
}
