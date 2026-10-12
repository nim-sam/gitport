package server

import (
	"errors"
	"sync"
	"time"

	"gorm.io/gorm"

	"github.com/nim-sam/gitport/internal/model"
	"github.com/nim-sam/gitport/internal/repository"
)

// activityMu serializes the read-modify-write of a user's activity row so that
// concurrent git sessions cannot lose an increment.
var activityMu sync.Mutex

// recordActivity applies mutate to the user's activity record, creating it on
// first use. It performs a single indexed lookup and one upsert against a table
// holding one row per user, so it never scans the (much larger) log table.
func recordActivity(userID uint, mutate func(*model.Activity)) {
	if userID == 0 {
		return
	}

	activityMu.Lock()
	defer activityMu.Unlock()

	a, err := repository.ActivityRepo.First("user_id = ?", userID)
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		a = model.Activity{UserID: userID}
	case err != nil:
		return
	}

	mutate(&a)
	repository.ActivityRepo.Save(a)
}

// recordLogin counts an interactive session opened by a resolved user.
func recordLogin(user model.User) {
	recordActivity(user.ID, func(a *model.Activity) {
		a.Logins++
		a.LastLogin = time.Now()
	})
}

// recordPush counts a successful push by a resolved user.
func recordPush(user model.User) {
	recordActivity(user.ID, func(a *model.Activity) {
		a.Pushes++
		a.LastPush = time.Now()
	})
}

// recordFetch counts a successful clone or pull by a resolved user.
func recordFetch(user model.User) {
	recordActivity(user.ID, func(a *model.Activity) {
		a.Fetches++
		a.LastFetch = time.Now()
	})
}
