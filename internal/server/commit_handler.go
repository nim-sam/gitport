package server

import (
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/nim-sam/gitport/internal/model"
	"github.com/nim-sam/gitport/internal/repository"
)

// InitCommits syncs the git history into the commits table. It is called by
// `gitport start` once the database is open, so that sessions and selection
// changes never have to walk git or recompute patches.
//
// Stores every commit reachable from HEAD in the database together with its
// diff. When the table already holds records only the missing commits are
// inserted, and records whose Diff is empty (e.g. rows created before the
// column existed) get their diff backfilled.
func (s Server) InitCommits() {
	s.syncCommits(s.Repo)
}

// RefreshCommits rebuilds the commits table from the served (bare) repository.
// It is called after a successful push: the table is cleared first so history
// that was rewritten (force push) does not leave unreachable rows behind, then
// re-populated from the current HEAD.
func (s Server) RefreshCommits() {
	repo, err := git.PlainOpen(s.bareRepoPath())
	if err != nil {
		return
	}
	repository.CommitRepo.Clear()
	s.syncCommits(repo)
}

// syncCommits is the shared core of InitCommits and RefreshCommits: it walks
// the commits reachable from repo's HEAD and upserts any that are new or whose
// cached diff is missing.
func (s Server) syncCommits(repo *git.Repository) {
	if repo == nil {
		return
	}

	byHash := make(map[string]model.Commit)
	for _, rec := range repository.CommitRepo.Load() {
		byHash[rec.Hash] = rec
	}

	gitCommits := loadCommits(repo, commitLimit)

	// Iterate oldest first so row ids ascend with commit order; the UI reads
	// commits back with id DESC, which then matches git log order even after
	// later restarts append newer commits.
	for i := len(gitCommits) - 1; i >= 0; i-- {
		c := gitCommits[i]

		if rec, ok := byHash[c.Hash]; ok {
			if rec.Diff == "" {
				rec.Diff = commitDiff(repo, c.Hash)
				repository.CommitRepo.Save(rec)
			}
			continue
		}

		c.Diff = commitDiff(repo, c.Hash)
		repository.CommitRepo.Save(c)
	}
}

// loadCommits fetches the <limit> most recent commits from the given git
// repository and returns a commit array. A repository without a git history,
// or without a git repository at all, yields an empty slice rather than an
// error.
func loadCommits(repo *git.Repository, limit int) []model.Commit {

	if repo == nil {
		return make([]model.Commit, 0)
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
		return make([]model.Commit, 0)
	}

	var out []model.Commit
	err = commitIter.ForEach(func(c *object.Commit) error {
		message := strings.TrimSpace(c.Message)

		out = append(out,
			model.Commit{
				Author:  c.Author.Name,
				Hash:    c.Hash.String(),
				Message: message,
				Date:    c.Author.When,
			})
		return nil
	})

	return out
}

// commitDiff computes the unified diff of the commit with hash <hash> straight
// from git. It is the producer half of the commit cache: syncCommits stores its
// result on model.Commit.Diff, and the UI renders that stored value rather than
// calling this on selection changes.
func commitDiff(repo *git.Repository, hash string) string {
	if repo == nil {
		return ""
	}

	h := plumbing.NewHash(hash)
	commit, err := repo.CommitObject(h)
	if err != nil {
		return ""
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
