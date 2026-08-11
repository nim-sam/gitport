package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/nim-sam/gitport/internal/auth"
	"github.com/nim-sam/gitport/ui"
)

func main() {
	userListTest := [...]*auth.User{
		auth.NewUser().
			WithName("Alice").
			WithSshKey("qwerqwerqwer").
			WithAccessLevel(auth.ReadWrite),
		auth.NewUser().
			WithName("Bob").
			WithSshKey("aadskljfa;lkjf").
			WithAccessLevel(auth.ReadWrite),
		auth.NewUser().
			WithName("Charlie").
			WithSshKey("pweaseletmein").
			WithAccessLevel(auth.ReadOnly),
	}

	d := ui.NewDashboard(userListTest[:])

	p := tea.NewProgram(d, tea.WithAltScreen())
	p.Run()

}
