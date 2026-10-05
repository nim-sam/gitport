package main

import (
	_ "strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nim-sam/gitport/internal/repository"
	"github.com/nim-sam/gitport/internal/server"
	"github.com/nim-sam/gitport/ui/ssh_tui/dashboard"
)

func main() {

	// fmt.Printf(server.Root)
	// fmt.Printf(server.LoadCommits(1)[0].Message)

	// commits := Server.LoadCommits(30)
	// println(Server.GetCommitDiff(commits[2].Hash))

	users := repository.UserRepo.Load()
	commits := server.GitportServer.LoadCommits(30)
	logs := repository.LogRepo.Load()

	d := dashboard.New(users, commits, logs)
	p := tea.NewProgram(d, tea.WithAltScreen())
	p.Run()

}
