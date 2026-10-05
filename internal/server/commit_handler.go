package server

import (
	"io"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/nim-sam/gitport/internal/model"
)

// Fetches <limit>th most recent commits from git server and returns
// commit array
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
				Hash:    c.Hash.String(),
				Message: message,
				Date:    c.Author.When.Format("Jan 02 2006"),
			})
		count++
		return nil
	})

	return out
}

// Fetches commit diff string for commit with hash <hash>
func (s Server) GetCommitDiff(hash string) string {
	h := plumbing.NewHash(hash)
	commit, err := s.Repo.CommitObject(h)
	if err != nil {
		panic(err)
	}

	parent, err := commit.Parent(0)
	var patch *object.Patch
	if err != nil {
		// First commit, diff against empty tree
		patch, _ = commit.Patch(nil)
	} else {
		patch, _ = parent.Patch(commit)
	}

	if patch == nil {
		return "No changes found."
	}

	return patch.String()
}
