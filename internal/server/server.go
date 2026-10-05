package server

import "github.com/go-git/go-git/v5"

type Server struct {
	Root      string
	Name      string
	Repo      *git.Repository
	ConfigDir string
}
