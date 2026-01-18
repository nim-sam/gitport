package tui

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nim-sam/gitport/pkg/logger"
)

type LogItem struct {
	level, desc, time string
}

func (i LogItem) FilterValue() string { return i.desc }

type logModel struct {
	list  list.Model
	ready bool
}

func (m logModel) Update(msg tea.Msg) (logModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		height := msg.Height - 1 // Leave room for footer
		if height < 1 {
			height = 1
		}
		m.list.SetSize(msg.Width, height)
		return m, nil
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m logModel) View() string {
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#505050"))
	pagination := helpStyle.Render(m.list.Paginator.View())
	help := helpStyle.Render("[up/down] Navigate logs  [tab] Switch tab")
	footer := lipgloss.JoinHorizontal(lipgloss.Left, pagination, "  ", help)

	return lipgloss.JoinVertical(lipgloss.Left, m.list.View(), footer)
}

type logDelegate struct{}

func (d logDelegate) Height() int                               { return 1 }
func (d logDelegate) Spacing() int                              { return 0 }
func (d logDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }

func (d logDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(LogItem)
	if !ok {
		return
	}

	// Column Widths
	levelWidth, timeWidth := 10, 20
	descWidth := m.Width() - levelWidth - timeWidth - 2

	// Level Styling
	levelStyle := lipgloss.NewStyle().Width(levelWidth).Padding(0, 1)
	switch i.level {
	case "ERROR":
		levelStyle = levelStyle.Foreground(lipgloss.Color("#FF1B1C")).Bold(true)
	case "WARN":
		levelStyle = levelStyle.Foreground(lipgloss.Color("#FFA500")).Bold(true)
	default:
		levelStyle = levelStyle.Foreground(lipgloss.Color("#909090"))
	}

	timeStyle := lipgloss.NewStyle().Width(timeWidth).Foreground(lipgloss.Color("242"))
	descStyle := lipgloss.NewStyle().Width(descWidth)

	// Row Highlight
	rowStr := lipgloss.JoinHorizontal(lipgloss.Top, levelStyle.Render(i.level), timeStyle.Render(i.time), descStyle.Render(i.desc))
	if index == m.Index() {
		fmt.Fprint(w, lipgloss.NewStyle().Background(lipgloss.Color("#5000ff")).Foreground(lipgloss.Color("#FFFFFF")).Render(rowStr))
	} else {
		fmt.Fprint(w, rowStr)
	}
}

func fetchLogItems() []list.Item {
	// Call the function we created in the logger package
	records, err := logger.ReadLogs()
	if err != nil {
		// Return a single error item if the file can't be read
		return []list.Item{LogItem{level: "ERROR", desc: "Could not read logs: " + err.Error(), time: ""}}
	}

	var items []list.Item
	// Skip the first row if it's the header "Date,Time,Level,Message"
	startIdx := 0
	if len(records) > 0 && records[0][0] == "Date" {
		startIdx = 1
	}

	for i := startIdx; i < len(records); i++ {
		row := records[i]
		// Ensure the row has enough columns to prevent index out of range
		if len(row) < 4 {
			continue
		}

		items = append(items, LogItem{
			time:  fmt.Sprintf("%s %s", row[0], row[1]), // Combines Date and Time
			level: row[2],
			desc:  row[3],
		})
	}

	// Optional: Reverse items if you want the newest logs at the top
	for i, j := 0, len(items)-1; i < j; i, j = i+1, j-1 {
		items[i], items[j] = items[j], items[i]
	}

	return items
}
