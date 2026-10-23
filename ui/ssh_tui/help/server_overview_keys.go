package help

import (
	"github.com/charmbracelet/bubbles/key"
)

// Help keys for the server overview tab

type ServerOverviewKeys struct {
	VNav key.Binding
	Tab  key.Binding
	Edit key.Binding
	Quit key.Binding
}

func (k ServerOverviewKeys) ShortHelp() []key.Binding {
	return []key.Binding{k.VNav, k.Tab, k.Edit, k.Quit}
}

func (k ServerOverviewKeys) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.VNav, k.Tab},
		{k.Edit},
		{k.Quit},
	}
}

var serverOverviewKeysAdmin = ServerOverviewKeys{
	VNav: key.NewBinding(
		key.WithKeys("up", "down"),
		key.WithHelp("↑/↓", "up/down"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "switch tabs"),
	),
	Edit: key.NewBinding(
		key.WithKeys("ctrl+e"),
		key.WithHelp("ctrl+e", "edit config"),
	),
	Quit: key.NewBinding(
		key.WithKeys("ctrl+q"),
		key.WithHelp("ctrl+q", "quit"),
	),
}

var serverOverviewKeysNonAdmin = ServerOverviewKeys{
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
