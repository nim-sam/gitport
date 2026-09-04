package config

import (
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/nim-sam/gitport/internal/server"
)

// Init server
var cwd, _ = os.Getwd()
var repo, _ = git.PlainOpen(cwd)
var GitportServer = server.Server{
	Root: cwd,
	Repo: repo,
}
