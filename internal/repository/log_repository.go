package repository

import (
	"github.com/nim-sam/gitport/internal/config"
	"github.com/nim-sam/gitport/internal/model"
	orm "github.com/nim-sam/gitport/pkg/mini_orm"
)

// Logs Repository
var LogRepo = orm.Repository[model.Log]{
	DB: config.DB,
}

func init() {
	LogRepo.DB.AutoMigrate(&model.Log{})
}
