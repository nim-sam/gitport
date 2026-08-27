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
	// userListTest := [...]auth.User{
	// 	auth.User{
	// 		Name:        "Alice",
	// 		Email:       "alice@mail.com",
	// 		SshKey:      "qwerqwerqwer",
	// 		AccessLevel: auth.ReadWrite,
	// 	},
	// 	auth.User{
	// 		Name:        "Bob",
	// 		Email:       "bubblybob@mail.com",
	// 		SshKey:      "aadskljfa;lkjf",
	// 		AccessLevel: auth.ReadWrite,
	// 	},
	// 	auth.User{
	// 		Name:        "Charlie",
	// 		Email:       "goodluckcharlie@gmail.com",
	// 		SshKey:      "pweaseletmein",
	// 		AccessLevel: auth.ReadOnly,
	// 	},
	// 	auth.User{
	// 		Name:        "Dalia",
	// 		Email:       "dalia@gmail.com",
	// 		SshKey:      "p123123412341234",
	// 		AccessLevel: auth.ReadOnly,
	// 	},
	// 	auth.User{
	// 		Name:        "Esme",
	// 		Email:       "esmerald@mail.com",
	// 		SshKey:      "df2q3ewr2",
	// 		AccessLevel: auth.Admin,
	// 	},
	// }

	// auth.UserRepo.Save(
	// 	auth.User{
	// 		Name:        "Nadeem",
	// 		Email:       "Nadeem@mail.com",
	// 		SshKey:      "Key",
	// 		AccessLevel: 0,
	// 	},
	// )

	userListTest := auth.UserRepo.Load()

	commitHistoryTest := [...]git.Commit{
		git.Commit{
			Author:  "Alice",
			Hash:    "abcdefg",
			Message: "First Commit",
			Date:    "today",
			Branch:  "remote/master",
		},
		git.Commit{
			Author:  "Bob",
			Hash:    "1234567",
			Message: "Second Commit",
			Date:    "july 7th 2026",
			Branch:  "remote/feature1",
		},
	}

	serverLogsTest := [...]logs.Log{
		logs.Log{
			Level:   logs.Warning,
			Date:    "2026/05/17 @ 16:00 UTC",
			Message: "The server is missing so and so information. Might compromise so and so features.",
		},
		logs.Log{
			Level:   logs.Error,
			Date:    "2026/05/17 @ 16:05 UTC",
			Message: "Somthing in the server failed. Whoops!",
		},
		logs.Log{
			Level:   logs.Info,
			Date:    "2026/05/17 @ 16:07 UTC",
			Message: "Server accepted connection from so and so user.",
		},
		logs.Log{
			Level:   logs.Warning,
			Date:    "2026/05/17 @ 16:00 UTC",
			Message: "The server is missing so and so information. Might compromise so and so features.",
		},
		logs.Log{
			Level:   logs.Error,
			Date:    "2026/05/17 @ 16:05 UTC",
			Message: "Somthing in the server failed. Whoops!",
		},
		logs.Log{
			Level:   logs.Info,
			Date:    "2026/05/17 @ 16:07 UTC",
			Message: "Server accepted connection from so and so user.",
		},
		logs.Log{
			Level:   logs.Warning,
			Date:    "2026/05/17 @ 16:00 UTC",
			Message: "The server is missing so and so information. Might compromise so and so features.",
		},
		logs.Log{
			Level:   logs.Error,
			Date:    "2026/05/17 @ 16:05 UTC",
			Message: "Somthing in the server failed. Whoops!",
		},
		logs.Log{
			Level:   logs.Info,
			Date:    "2026/05/17 @ 16:07 UTC",
			Message: "Server accepted connection from so and so user.",
		},
		logs.Log{
			Level:   logs.Warning,
			Date:    "2026/05/17 @ 16:00 UTC",
			Message: "The server is missing so and so information. Might compromise so and so features.",
		},
		logs.Log{
			Level:   logs.Error,
			Date:    "2026/05/17 @ 16:05 UTC",
			Message: "Somthing in the server failed. Whoops!",
		},
		logs.Log{
			Level:   logs.Info,
			Date:    "2026/05/17 @ 16:07 UTC",
			Message: "Server accepted connection from so and so user.",
		},
	}

	_ = serverLogsTest

	d := ssh_tui.NewDashboard(userListTest[:], commitHistoryTest[:], serverLogsTest[:])

	p := tea.NewProgram(d, tea.WithAltScreen())
	p.Run()

}
