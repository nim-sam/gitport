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
	Width  int
	Height int
}

func (s Model) Init() tea.Cmd { return nil }
func (s Model) Update(msg tea.Msg) (Model, tea.Cmd) {

	// Updating column widths
	c := s.table.Columns()
	c[0].Width = 10
	c[1].Width = 25
	c[2].Width = s.Width - (c[0].Width + c[1].Width) - 4
	s.table.SetColumns(c)

	// General updates
	s.table, _ = s.table.Update(msg)
	s.table.SetWidth(s.Width - 4)
	s.table.SetHeight(s.Height - 2)
	return s, nil
}
func (s Model) View() string {

	s.table.SetStyles(table.Styles{
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
	return tableContainer.Render(s.table.View())
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
