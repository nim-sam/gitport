package repository

import (
	"github.com/nim-sam/gitport/internal/config"
	"github.com/nim-sam/gitport/internal/model"
	"github.com/nim-sam/gitport/pkg/mini_orm"
)

// User Repository
var UserRepo = mini_orm.Repository[model.User]{
	DB: config.DB,
}

func init() {
	UserRepo.DB.AutoMigrate(&model.User{})
}
