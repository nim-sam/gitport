package serveroverview

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/nim-sam/gitport/ui"
)

type Model struct {
	width  int
	height int
}

func New() Model {
	return Model{}
}

func (m *Model) Width(w int)  { m.width = w }
func (m *Model) Height(h int) { m.height = h }

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Msg) {
	return m, nil
}

func (m Model) View() string {

	overviewTable := table.New().Border(
		lipgloss.RoundedBorder(),
	).BorderStyle(
		lipgloss.NewStyle().Foreground(ui.Palette.Border),
	).BorderRow(
		true,
	).Width(
		m.width+2,
	).Height(
		m.height,
	).Rows(
		[]string{
			"Hostname", "name-of-the-server",
		},
		[]string{
			"Binding Address", "123.123.123.123",
		},
		[]string{
			"Port", "2222",
		},
	)

	return overviewTable.Render()
}
