// Package cli wires the GitPort commands (init, start) through the pecan
// command registry and dispatches on the process arguments.
package cli

import (
	"fmt"
	"os"

	"github.com/nim-sam/pecan"
)

const usage = `GitPort is a self-hosted, ssh-accessible git server dashboard.

Usage:
  gitport init    Prompt for the server's starter configuration, then create
                  the .gitport directory with the config and database.
  gitport start   Run the server. Requires an initialized repository.
`

// Run registers the command tree and dispatches on os.Args[1:].
func Run() error {

	pecan.NewCommand().
		WithName("").
		WithDesc("GitPort is a self-hosted, ssh-accessible git server dashboard.").
		WithUsage("gitport <command>").
		WithHandler(rootHandler).
		Register()

	// pecan.NewCommand().
	// 	WithName("init").
	// 	WithDesc("Initialize a repository as a GitPort server.").
	// 	WithUsage("gitport init").
	// 	WithHandler(initHandler).
	// 	Register()

	pecan.NewCommand().
		WithName("start").
		WithDesc("Run the GitPort server.").
		WithUsage("gitport start").
		WithHandler(startHandler).
		Register()

	if _, err := pecan.App("gitport").Launch(); err != nil {
		fmt.Fprint(os.Stderr, usage)
		return err
	}
	return nil
}

func rootHandler(_ []string) error {
	fmt.Print(usage)
	return nil
}
