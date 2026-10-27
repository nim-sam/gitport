package serverlogs

import (
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nim-sam/gitport/internal/model"
	"github.com/nim-sam/gitport/ui"
)

type Model struct {
	table  table.Model
	width  int
	height int
}

// Dimension setters
func (m *Model) Width(width int)   { m.width = width }
func (m *Model) Height(height int) { m.height = height }

func (m Model) Init() tea.Cmd { return nil }
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {

	// Updating column widths
	c := m.table.Columns()
	c[0].Width = 10
	c[1].Width = 25
	c[2].Width = m.width - (c[0].Width + c[1].Width) - 4
	m.table.SetColumns(c)

	// General updates
	m.table, _ = m.table.Update(msg)
	m.table.SetWidth(m.width - 4)
	m.table.SetHeight(m.height - 2)
	return m, nil
}
func (m Model) View() string {

	m.table.SetStyles(table.Styles{
		Selected: lipgloss.NewStyle().
			Foreground(ui.Palette.Accent).
			Bold(true),
		Header: lipgloss.NewStyle().
			Foreground(ui.Palette.FgHighlight).
			Bold(true).
			Border(lipgloss.MarkdownBorder(), false, false, true, false).
			BorderForeground(ui.Palette.SelectedBorder),
	})

	tableContainer := lipgloss.NewStyle().AlignHorizontal(lipgloss.Left).Padding(0, 2)
	return tableContainer.Render(m.table.View())
}

func New(l []model.Log) Model {

	cols := []table.Column{
		{Title: "LEVEL", Width: 10},
		{Title: "DATE", Width: 25},
		{Title: "MESSAGE", Width: 30},
	}

	var rows []table.Row
	for _, log := range l {
		var logLevel string
		switch log.Level {
		case model.Warning:
			logLevel = "Warning"
		case model.Error:
			logLevel = "Error"
		case model.Info:
			logLevel = "Info"
		}

		rows = append([]table.Row{[]string{logLevel, log.Date, log.Message}}, rows...)
		// rows = append(rows, []string{logLevel, log.Date, log.Message})
	}

	newTable := table.New(
		table.WithColumns(cols),
		table.WithRows(rows),
		table.WithFocused(true),
	)

	return Model{
		table: newTable,
	}
}
