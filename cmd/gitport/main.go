package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/nim-sam/gitport/internal/auth"
	_ "github.com/nim-sam/gitport/internal/config"
	"github.com/nim-sam/gitport/internal/git"
	"github.com/nim-sam/gitport/internal/logs"
	"github.com/nim-sam/gitport/ui/ssh_tui"
)

func main() {

	users := auth.UserRepo.Load()
	commitHistory := git.CommitRepo.Load()
	serverLogs := logs.LogRepo.Load()

	d := ssh_tui.NewDashboard(users[:], commitHistory[:], serverLogs[:])

	p := tea.NewProgram(d, tea.WithAltScreen())
	p.Run()

}
