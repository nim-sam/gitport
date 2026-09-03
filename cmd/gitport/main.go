package main

import (
	"os"
	_ "strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nim-sam/gitport/internal/server"
	"github.com/nim-sam/gitport/ui/ssh_tui"
)

func main() {

	// Init server
	cwd, _ := os.Getwd()
	server := server.Server{
		Root: cwd,
	}

	// fmt.Printf(server.Root)
	// fmt.Printf(server.LoadCommits(1)[0].Message)

	d := ssh_tui.NewDashboard(server)
	p := tea.NewProgram(d, tea.WithAltScreen())
	p.Run()

}
