package repository

import (
	"github.com/nim-sam/gitport/internal/config"
	"github.com/nim-sam/gitport/internal/model"
	localorm "github.com/nim-sam/gitport/pkg/mini_orm"
)

// Commit Repository
var CommitRepo = localorm.Repository[model.Commit]{
	DB: config.DB,
}

func init() {
	CommitRepo.DB.AutoMigrate(&model.Commit{})
}
