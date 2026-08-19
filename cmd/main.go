package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/nim-sam/gitport/internal/auth"
	"github.com/nim-sam/gitport/internal/git"
	"github.com/nim-sam/gitport/internal/logs"
	"github.com/nim-sam/gitport/ui"
)

func main() {
	userListTest := [...]*auth.User{
		auth.NewUser().
			WithName("Alice").
			WithEmail("alice@mail.com").
			WithSshKey("qwerqwerqwer").
			WithAccessLevel(auth.ReadWrite),
		auth.NewUser().
			WithName("Bob").
			WithEmail("bubblybob@mail.com").
			WithSshKey("aadskljfa;lkjf").
			WithAccessLevel(auth.ReadWrite),
		auth.NewUser().
			WithName("Charlie").
			WithEmail("goodluckcharlie@gmail.com").
			WithSshKey("pweaseletmein").
			WithAccessLevel(auth.ReadOnly),
		auth.NewUser().
			WithName("Dalia").
			WithEmail("dalia@gmail.com").
			WithSshKey("p123123412341234").
			WithAccessLevel(auth.ReadOnly),
		auth.NewUser().
			WithName("Esme").
			WithEmail("esmerald@mail.com").
			WithSshKey("df2q3ewr2").
			WithAccessLevel(auth.Admin),
	}

	commitHistoryTest := [...]*git.Commit{
		git.NewCommit().
			WithAuthor("Alice").
			WithHash("abcdefg").
			WithMessage("First Commit").
			WithDate("today").
			WithBranch("remote/master"),
		git.NewCommit().
			WithAuthor("Bob").
			WithHash("1234567").
			WithMessage("Second Commit").
			WithDate("july 7th 2026").
			WithBranch("remote/feature1"),
	}

	serverLogsTest := [...]*logs.Log{
		logs.NewLog().
			WithLevel(logs.Warning).
			WithDate("2026/05/17 @ 16:00 UTC").
			WithMessage("The server is missing so and so information. Might compromise so and so features."),
		logs.NewLog().
			WithLevel(logs.Error).
			WithDate("2026/05/17 @ 16:05 UTC").
			WithMessage("Somthing in the server failed. Whoops!"),
		logs.NewLog().
			WithLevel(logs.Info).
			WithDate("2026/05/17 @ 16:07 UTC").
			WithMessage("Server accepted connection from so and so user."),
		logs.NewLog().
			WithLevel(logs.Warning).
			WithDate("2026/05/17 @ 16:00 UTC").
			WithMessage("The server is missing so and so information. Might compromise so and so features."),
		logs.NewLog().
			WithLevel(logs.Error).
			WithDate("2026/05/17 @ 16:05 UTC").
			WithMessage("Somthing in the server failed. Whoops!"),
		logs.NewLog().
			WithLevel(logs.Info).
			WithDate("2026/05/17 @ 16:07 UTC").
			WithMessage("Server accepted connection from so and so user."),
		logs.NewLog().
			WithLevel(logs.Warning).
			WithDate("2026/05/17 @ 16:00 UTC").
			WithMessage("The server is missing so and so information. Might compromise so and so features."),
		logs.NewLog().
			WithLevel(logs.Error).
			WithDate("2026/05/17 @ 16:05 UTC").
			WithMessage("Somthing in the server failed. Whoops!"),
		logs.NewLog().
			WithLevel(logs.Info).
			WithDate("2026/05/17 @ 16:07 UTC").
			WithMessage("Server accepted connection from so and so user."),
		logs.NewLog().
			WithLevel(logs.Warning).
			WithDate("2026/05/17 @ 16:00 UTC").
			WithMessage("The server is missing so and so information. Might compromise so and so features."),
		logs.NewLog().
			WithLevel(logs.Error).
			WithDate("2026/05/17 @ 16:05 UTC").
			WithMessage("Somthing in the server failed. Whoops!"),
		logs.NewLog().
			WithLevel(logs.Info).
			WithDate("2026/05/17 @ 16:07 UTC").
			WithMessage("Server accepted connection from so and so user."),
		logs.NewLog().
			WithLevel(logs.Warning).
			WithDate("2026/05/17 @ 16:00 UTC").
			WithMessage("The server is missing so and so information. Might compromise so and so features."),
		logs.NewLog().
			WithLevel(logs.Error).
			WithDate("2026/05/17 @ 16:05 UTC").
			WithMessage("Somthing in the server failed. Whoops!"),
		logs.NewLog().
			WithLevel(logs.Info).
			WithDate("2026/05/17 @ 16:07 UTC").
			WithMessage("Server accepted connection from so and so user."),
		logs.NewLog().
			WithLevel(logs.Warning).
			WithDate("2026/05/17 @ 16:00 UTC").
			WithMessage("The server is missing so and so information. Might compromise so and so features."),
		logs.NewLog().
			WithLevel(logs.Error).
			WithDate("2026/05/17 @ 16:05 UTC").
			WithMessage("Somthing in the server failed. Whoops!"),
		logs.NewLog().
			WithLevel(logs.Info).
			WithDate("2026/05/17 @ 16:07 UTC").
			WithMessage("Server accepted connection from so and so user."),
		logs.NewLog().
			WithLevel(logs.Warning).
			WithDate("2026/05/17 @ 16:00 UTC").
			WithMessage("The server is missing so and so information. Might compromise so and so features."),
		logs.NewLog().
			WithLevel(logs.Error).
			WithDate("2026/05/17 @ 16:05 UTC").
			WithMessage("Somthing in the server failed. Whoops!"),
		logs.NewLog().
			WithLevel(logs.Info).
			WithDate("2026/05/17 @ 16:07 UTC").
			WithMessage("Server accepted connection from so and so user."),
	}

	_ = serverLogsTest

	d := ui.NewDashboard(userListTest[:], commitHistoryTest[:], serverLogsTest[:])

	p := tea.NewProgram(d, tea.WithAltScreen())
	p.Run()

}
