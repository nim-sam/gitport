package model

import (
	"time"

	"gorm.io/gorm"
)

// Struct holding a user's cumulative activity. It is kept in its own table
// rather than on User because editing a user saves the whole form-shaped User,
// which would wipe any counters stored alongside it.
type Activity struct {
	gorm.Model
	UserID    uint `gorm:"uniqueIndex"`
	LastLogin time.Time
	LastPush  time.Time
	LastFetch time.Time
	Logins    int
	Pushes    int
	Fetches   int
}
