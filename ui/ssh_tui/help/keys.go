package help

import "github.com/charmbracelet/bubbles/key"

type DashboardKeys struct {
	VNav       key.Binding
	Tab        key.Binding
	NewUser    key.Binding
	EditUser   key.Binding
	DeleteUser key.Binding
	Quit       key.Binding
	Help       key.Binding
}

func (d DashboardKeys) ShortHelp() []key.Binding {
	return []key.Binding{d.VNav, d.Tab, d.Quit, d.Help}
}

func (d DashboardKeys) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{d.VNav, d.Tab},
		{d.NewUser, d.EditUser},
		{d.DeleteUser},
		{d.Quit, d.Help},
	}
}

var keys = DashboardKeys{
	VNav: key.NewBinding(
		key.WithKeys("up", "down"),
		key.WithHelp("↑/↓", "up/down"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "switch tabs"),
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
