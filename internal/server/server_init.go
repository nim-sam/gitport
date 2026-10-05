package server

import (
	"os"

	"github.com/go-git/go-git/v5"
)

// Init server
var cwd, _ = os.Getwd()
var repo, _ = git.PlainOpen(cwd)
var configDir, _ = os.UserConfigDir()

// GitportServer is the process-wide server singleton.
//
// Bind and Port are populated from command line flags in main and must
// therefore be non-empty before Start is called: an empty port makes
// net.Listen fall back to an OS-assigned ephemeral port, so the server would
// listen somewhere other than where it advertises.
var GitportServer = Server{
	Root:      cwd,
	Repo:      repo,
	ConfigDir: configDir,
}
