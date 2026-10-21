package help

import (
	"github.com/charmbracelet/bubbles/key"
)

// Dashboard Keys interface that key binding structs will implement for each
// available dashboard tab
type DashboardKeys interface {
	ShortHelp() []key.Binding
	FullHelp() [][]key.Binding
}
