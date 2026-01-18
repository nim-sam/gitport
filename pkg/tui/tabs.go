package main

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type sessionState int

const (
	tabCommits sessionState = iota
	tabBranches
	tabSettings
)

//var tabs = []string{"Commits", "Branches", "Settings"}

type mainModel struct {
	state     sessionState
	activeTab int
	commitLog commitModel // Your existing model
	width     int
	height    int
}

func (m mainModel) Init() tea.Cmd {
	return nil
}

func (m mainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// IMPORTANT: Always pass size to sub-models so they can initialize
		// Subtract a bit of height for your tab header
		subMsg := tea.WindowSizeMsg{Width: msg.Width, Height: msg.Height - 4}
		var newModel tea.Model
		newModel, cmd = m.commitLog.Update(subMsg)
		m.commitLog = newModel.(commitModel)
		return m, cmd // Or append to cmds

	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			m.activeTab = (m.activeTab + 1) % 3
			return m, nil
		case "ctrl+c":
			return m, tea.Quit
		}
	}

	// Route regular messages (keys, etc.) only to the active tab
	if m.activeTab == 1 {
		var newModel tea.Model
		newModel, cmd = m.commitLog.Update(msg)
		m.commitLog = newModel.(commitModel)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m mainModel) View() string {
	doc := strings.Builder{}

	// Tab Styling
	tabNames := []string{"Dashboard", "Commit History", "Logs"}
	var tabs []string
	for i, name := range tabNames {
		style := lipgloss.NewStyle().Padding(0, 1).MarginRight(1)
		if m.activeTab == i {
			style = style.Foreground(lipgloss.Color("#FFFFFF")).
				Background(lipgloss.Color("#5000ff")).Bold(true)
		} else {
			style = style.Foreground(lipgloss.Color("240"))
		}
		tabs = append(tabs, style.Render(name))
	}

	// Header bar
	header := lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
	doc.WriteString(header + "\n\n")

	// Content Switching
	switch m.activeTab {

	case 0:
		doc.WriteString("Dashboard coming soon...")
	case 1:
		doc.WriteString(m.commitLog.View())
	case 2:
		doc.WriteString("Logs coming soon...")

	}

	return doc.String()
}
