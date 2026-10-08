package server

import (
	// "io"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/nim-sam/gitport/internal/model"
	"github.com/nim-sam/gitport/internal/repository"
)

// Syncs the git history into the commits table at process start so that
// sessions and selection changes never have to walk git or recompute patches.
// Runs from init(), after the database, repository migrations and the package
// level Server singleton have been initialized.
func init() {
	GitportServer.InitCommits()
}

// Stores every commit reachable from HEAD in the database together with its
// diff. When the table already holds records only the missing commits are
// inserted, and records whose Diff is empty (e.g. rows created before the
// column existed) get their diff backfilled.
func (s Server) InitCommits() {
	byHash := make(map[string]model.Commit)
	for _, rec := range repository.CommitRepo.Load() {
		byHash[rec.Hash] = rec
	}

	gitCommits := s.LoadCommits(commitLimit)

	// Iterate oldest first so row ids ascend with commit order; the UI reads
	// commits back with id DESC, which then matches git log order even after
	// later restarts append newer commits.
	for i := len(gitCommits) - 1; i >= 0; i-- {
		c := gitCommits[i]

		if rec, ok := byHash[c.Hash]; ok {
			if rec.Diff == "" {
				rec.Diff = s.GetCommitDiff(c.Hash)
				repository.CommitRepo.Save(rec)
			}
			continue
		}

		c.Diff = s.GetCommitDiff(c.Hash)
		repository.CommitRepo.Save(c)
	}
}

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

	// Iterate through repo commits and return commit array. Committer-time
	// order matches git log; the default depth-first traversal interleaves
	// merged branches and yields commits out of date order.
	commitIter, err := repo.Log(&git.LogOptions{From: ref.Hash(), Order: git.LogOrderCommitterTime})
	if err != nil {
		panic(err)
	}

	var out []model.Commit
	// count := 0
	err = commitIter.ForEach(func(c *object.Commit) error {
		// if count >= limit {
		// 	return io.EOF
		// }

		message := strings.TrimSpace(c.Message)

		out = append(out,
			model.Commit{
				Author:  c.Author.Name,
				Hash:    c.Hash.String(),
				Message: message,
				Date:    c.Author.When.Format(model.DateLayout),
			})
		// count++
		return nil
	})

	return out
}

// Computes the unified diff of the commit with hash <hash> straight from git.
// It is the producer half of the commit cache: InitCommits stores its result
// on model.Commit.Diff, and the UI renders that stored value rather than
// calling this on selection changes.
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
