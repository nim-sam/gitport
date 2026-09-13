package repository

import (
	"github.com/nim-sam/gitport/internal/database"
	"github.com/nim-sam/gitport/internal/model"
	orm "github.com/nim-sam/gitport/pkg/mini_orm"
)

// Logs Repository
var LogRepo = orm.Repository[model.Log]{
	DB: database.DB,
}

func init() {
	LogRepo.DB.AutoMigrate(&model.Log{})
}
