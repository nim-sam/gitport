package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type sessionState int

//var tabs = []string{"Commits", "Branches", "Settings"}

type mainModel struct {
	state     sessionState
	activeTab int
	dashboard dashboardModel
	commitLog commitModel // Your existing model
	logFinder logModel
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

		// Propagate size to ALL models so they can calculate layout
		// Subtract height for your header (tabs + spacing)
		subMsg := tea.WindowSizeMsg{Width: m.width, Height: m.height - 3}

		var cmdD, cmdC, cmdL tea.Cmd
		m.dashboard, cmdD = m.dashboard.Update(subMsg)

		// If commitLog returns tea.Model, we assert it back
		newCommit, cmdC := m.commitLog.Update(subMsg)
		m.commitLog = newCommit.(commitModel)

		m.logFinder.list, cmdL = m.logFinder.list.Update(subMsg)

		return m, tea.Batch(cmdD, cmdC, cmdL)

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
	switch m.activeTab {
	case 0:
		// m.dashboard.Update returns (dashboardModel, tea.Cmd)
		m.dashboard, cmd = m.dashboard.Update(msg)
		cmds = append(cmds, cmd)
	case 1:
		// If your commitLog.Update returns (tea.Model, tea.Cmd)
		var newModel tea.Model
		newModel, cmd = m.commitLog.Update(msg)
		m.commitLog = newModel.(commitModel)
		cmds = append(cmds, cmd)
	case 2:
		m.logFinder.list, cmd = m.logFinder.list.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m mainModel) View() string {
	if m.width == 0 || m.height == 0 {
		return "Initializing..."
	}

	// 1. Render Tabs
	tabNames := []string{"Dashboard", "Commit History", "Logs"}
	var tabs []string
	for i, name := range tabNames {
		style := lipgloss.NewStyle().Padding(0, 2)
		if m.activeTab == i {
			style = style.Background(lipgloss.Color("#5000ff")).Foreground(lipgloss.Color("#FFFFFF"))
		}
		tabs = append(tabs, style.Render(name))
	}
	header := lipgloss.JoinHorizontal(lipgloss.Top, tabs...)

	// 2. Get Sub-view Content
	var content string
	switch m.activeTab {
	case 0:
		content = m.dashboard.View()
	case 1:
		content = m.commitLog.View()
	case 2:
		content = m.logFinder.list.View()
	}

	// 3. Join vertically and ensure no accidental wrapping
	// We use MaxHeight to prevent the TUI from "pushing" the terminal prompt down
	return lipgloss.NewStyle().
		Width(m.width).
		MaxHeight(m.height).
		Render(header + "\n\n" + content)
}
