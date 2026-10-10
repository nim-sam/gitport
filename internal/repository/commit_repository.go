package repository

import (
	"sort"
	"time"

	"github.com/nim-sam/gitport/internal/model"
	orm "github.com/nim-sam/gitport/pkg/mini_orm"
)

// Commit Repository
var CommitRepo = orm.Repository[model.Commit]{}

// LoadCommitsNewestFirst returns every stored commit, newest first: ordered
// by author date descending, ties broken by descending id (insertion order).
//
// The ordering is derived from the date itself rather than from id order
// because ids do not reliably reflect it: the git walk that populated the
// table historically traversed merged branches out of date order, and a row
// re-inserted after being deleted receives the next unused id regardless of
// its date. Rows whose Date cannot be parsed sort last.
func LoadCommitsNewestFirst() []model.Commit {
	var commits []model.Commit

	result := CommitRepo.DB.Find(&commits)
	if err := result.Error; err != nil {
		panic(err)
	}

	sort.Slice(commits, func(i, j int) bool {
		di, dj := commitDate(commits[i]), commitDate(commits[j])
		if !di.Equal(dj) {
			return di.After(dj)
		}
		return commits[i].ID > commits[j].ID
	})

	return commits
}

// commitDate parses a stored commit date, returning the zero time when the
// date is missing or malformed so such rows sink to the end of date-desc
// orderings.
func commitDate(c model.Commit) time.Time {
	t, err := time.Parse(model.DateLayout, c.Date)
	if err != nil {
		return time.Time{}
	}
	return t
}
