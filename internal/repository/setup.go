package repository

import (
	"github.com/nim-sam/gitport/internal/database"
	"github.com/nim-sam/gitport/internal/model"
)

// Migrate binds the open database handle to every repository and creates the
// schema. It must be called after database.Init: the package-level repository
// values capture database.DB while it is still nil, so they are wired up here
// rather than at import time.
func Migrate() {
	UserRepo.DB = database.DB
	LogRepo.DB = database.DB
	CommitRepo.DB = database.DB
	ActivityRepo.DB = database.DB

	UserRepo.DB.AutoMigrate(&model.User{})
	LogRepo.DB.AutoMigrate(&model.Log{})
	CommitRepo.DB.AutoMigrate(&model.Commit{})
	ActivityRepo.DB.AutoMigrate(&model.Activity{})
}
