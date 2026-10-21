package help

import (
	"github.com/charmbracelet/bubbles/key"
)

// Help keys for the user management tab

type CommitHistoryKeys struct {
	VNav         key.Binding
	Tab          key.Binding
	SwitchPanels key.Binding
	FuzzySearch  key.Binding
	Quit         key.Binding
	Help         key.Binding
}

func (k CommitHistoryKeys) ShortHelp() []key.Binding {
	return []key.Binding{k.VNav, k.Tab, k.Quit, k.Help}
}

func (k CommitHistoryKeys) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.VNav, k.Tab},
		{k.SwitchPanels, k.FuzzySearch},
		{k.Quit, k.Help},
	}
}

var commitHistoryKeys = CommitHistoryKeys{
	VNav: key.NewBinding(
		key.WithKeys("up", "down"),
		key.WithHelp("↑/↓", "up/down"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "switch tabs"),
	),
	SwitchPanels: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "Switch Panel Focus"),
	),
	FuzzySearch: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "Filter Commits"),
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
