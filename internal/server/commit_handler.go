package server

import (
	"io"
	"strings"

	_ "github.com/charmbracelet/wish"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/nim-sam/gitport/internal/model"
)

func (s Server) LoadCommits(limit int) []model.Commit {

	repo, err := git.PlainOpen(s.Root)
	if err != nil {
		panic(err)
	}

	ref, err := repo.Head()
	if err != nil {
		// If there is no head, i.e empty Repo, return empty commit array
		return make([]model.Commit, 0)
	}

	// Iterate through repo commits and return commit array
	commitIter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		panic(err)
	}

	var out []model.Commit
	count := 0
	err = commitIter.ForEach(func(c *object.Commit) error {
		if count >= limit {
			return io.EOF
		}

		message := strings.TrimSpace(c.Message)

		out = append(out,
			model.Commit{
				Author:  c.Author.Name,
				Hash:    c.Hash.String()[:6],
				Message: message,
				Date:    c.Author.When.Format("Jan 02 2006"),
			})
		count++
		return nil
	})

	return out
}
