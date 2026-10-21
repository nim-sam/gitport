package help

import (
	"github.com/charmbracelet/bubbles/key"
)

// Help keys for the user management tab

type ServerLogsKeys struct {
	VNav key.Binding
	Tab  key.Binding
	Quit key.Binding
}

func (k ServerLogsKeys) ShortHelp() []key.Binding {
	return []key.Binding{k.VNav, k.Tab, k.Quit}
}

func (k ServerLogsKeys) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.VNav, k.Tab, k.Quit}}
}

var serverLogsKeys = ServerLogsKeys{
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
