package repository

import (
	"github.com/nim-sam/gitport/internal/config"
	"github.com/nim-sam/gitport/internal/model"
	orm "github.com/nim-sam/gitport/pkg/mini_orm"
)

// Commit Repository
var CommitRepo = orm.Repository[model.Commit]{
	DB: config.DB,
}

func init() {
	CommitRepo.DB.AutoMigrate(&model.Commit{})
}
