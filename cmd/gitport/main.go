package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/nim-sam/gitport/internal/auth"
	"github.com/nim-sam/gitport/internal/git"
	"github.com/nim-sam/gitport/internal/logs"
	"github.com/nim-sam/gitport/ui/ssh_tui"
)

func main() {
	userListTest := [...]auth.User{
		auth.User{
			Name: "Alice",
			Email: "alice@mail.com",
			SshKey: "qwerqwerqwer",
			AccessLevel: auth.ReadWrite,
		},
		auth.User{
			Name: "Bob",
			Email: "bubblybob@mail.com",
			SshKey: "aadskljfa;lkjf",
			AccessLevel: auth.ReadWrite,
		},
		auth.User{
			Name: "Charlie",
			Email: "goodluckcharlie@gmail.com",
			SshKey: "pweaseletmein",
			AccessLevel: auth.ReadOnly,
		},
		auth.User{
			Name: "Dalia",
			Email: "dalia@gmail.com",
			SshKey: "p123123412341234",
			AccessLevel: auth.ReadOnly,
		},
		auth.User{
			Name: "Esme",
			Email: "esmerald@mail.com",
			SshKey: "df2q3ewr2",
			AccessLevel: auth.Admin,
		},
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

	d := ssh_tui.NewDashboard(userListTest[:], commitHistoryTest[:], serverLogsTest[:])

	p := tea.NewProgram(d, tea.WithAltScreen())
	p.Run()

}
