package main

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type LogItem struct {
	level, desc, time string
}

func (i LogItem) FilterValue() string { return i.desc }

type logModel struct {
	list  list.Model
	ready bool
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
