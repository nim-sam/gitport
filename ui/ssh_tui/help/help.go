package help

import (
	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	keys    DashboardKeys
	help    help.Model
	lastKey string
	Width   int
	Height  int
}

func New() Model {
	return Model{
		keys:   keys,
		help:   help.New(),
		Height: 2,
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
		case "?":
			m.help.ShowAll = !m.help.ShowAll
		}
	}

	return m, nil
}

func (m Model) View() string {

	helpContainer := lipgloss.NewStyle().Width(m.Width).Height(m.Height).AlignHorizontal(lipgloss.Center)

	return helpContainer.Render(m.help.View(m.keys))
}
