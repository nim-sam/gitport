package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/nim-sam/gitport/ui/ssh_tui"
)

func main() {
	d := ssh_tui.NewDashboard()
	p := tea.NewProgram(d, tea.WithAltScreen())
	p.Run()
}
