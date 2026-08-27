package main

import (
	tea "github.com/charmbracelet/bubbletea"
	_ "github.com/nim-sam/gitport/internal/config"
	"github.com/nim-sam/gitport/internal/repository"
	"github.com/nim-sam/gitport/ui/ssh_tui"
)

func main() {

	users := repository.UserRepo.Load()
	commitHistory := repository.CommitRepo.Load()
	serverLogs := repository.LogRepo.Load()

	d := ssh_tui.NewDashboard(users[:], commitHistory[:], serverLogs[:])

	p := tea.NewProgram(d, tea.WithAltScreen())
	p.Run()

}
