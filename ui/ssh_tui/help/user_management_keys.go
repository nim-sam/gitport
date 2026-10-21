package help

import (
	"github.com/charmbracelet/bubbles/key"
)

// Help keys for the user management tab

type UserManagementKeys struct {
	VNav       key.Binding
	Tab        key.Binding
	NewUser    key.Binding
	EditUser   key.Binding
	DeleteUser key.Binding
	Quit       key.Binding
	Help       key.Binding
}

func (k UserManagementKeys) ShortHelp() []key.Binding {
	return []key.Binding{k.VNav, k.Tab, k.Quit, k.Help}
}

func (k UserManagementKeys) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.VNav, k.Tab},
		{k.NewUser, k.EditUser},
		{k.DeleteUser},
		{k.Quit, k.Help},
	}
}

var userManagementKeysAdmin = UserManagementKeys{
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

var userManagementKeysNonAdmin = UserManagementKeys{
	VNav: key.NewBinding(
		key.WithKeys("up", "down"),
		key.WithHelp("↑/↓", "up/down"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "switch tabs"),
	),
	Quit: key.NewBinding(
		key.WithKeys("ctrl+q"),
		key.WithHelp("ctrl+q", "quit"),
	),
}
