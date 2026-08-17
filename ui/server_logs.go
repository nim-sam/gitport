package ui

import (
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nim-sam/gitport/internal/logs"
)

type ServerLogsModel struct {
	table  table.Model
	width  int
	height int
}

func (s ServerLogsModel) Init() tea.Cmd { return nil }
func (s ServerLogsModel) Update(msg tea.Msg) (ServerLogsModel, tea.Cmd) {
	s.table, _ = s.table.Update(msg)
	s.table.SetWidth(s.width)
	s.table.SetHeight(s.height)
	return s, nil
}
func (s ServerLogsModel) View() string {
	tableContainer := lipgloss.NewStyle().AlignHorizontal(lipgloss.Left)
	return tableContainer.Render(s.table.View())
}

func NewServerLogsModel(l []*logs.Log) ServerLogsModel {

	cols := []table.Column{
		{Title: "Level", Width: 10},
		{Title: "Date", Width: 25},
		{Title: "Message", Width: 30},
	}

	var rows []table.Row
	for _, log := range l {
		var logLevel string
		switch log.Level() {
		case logs.Warning:
			logLevel = "Warning"
		case logs.Error:
			logLevel = "Error"
		case logs.Info:
			logLevel = "Info"
		}
		rows = append(rows, []string{logLevel, log.Date(), log.Message()})
	}

	newTable := table.New(
		table.WithColumns(cols),
		table.WithRows(rows),
		table.WithFocused(true),
	)

	return ServerLogsModel{
		table: newTable,
	}
}
