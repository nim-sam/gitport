package main

import (
	_ "strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nim-sam/gitport/internal/config"
	"github.com/nim-sam/gitport/ui/ssh_tui"
)

func main() {

	// fmt.Printf(server.Root)
	// fmt.Printf(server.LoadCommits(1)[0].Message)

	// commits := Server.LoadCommits(30)
	// println(Server.GetCommitDiff(commits[2].Hash))

	d := ssh_tui.NewDashboard(config.GitportServer)
	p := tea.NewProgram(d, tea.WithAltScreen())
	p.Run()

}
