package repository

import (
	"github.com/nim-sam/gitport/internal/database"
	"github.com/nim-sam/gitport/internal/model"
	orm "github.com/nim-sam/gitport/pkg/mini_orm"
)

// Commit Repository
var CommitRepo = orm.Repository[model.Commit]{
	DB: database.DB,
}

func init() {
	CommitRepo.DB.AutoMigrate(&model.Commit{})
}
