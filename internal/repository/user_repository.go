package repository

import (
	"github.com/nim-sam/gitport/internal/config"
	"github.com/nim-sam/gitport/internal/model"
	orm "github.com/nim-sam/gitport/pkg/mini_orm"
)

// User Repository
var UserRepo = orm.Repository[model.User]{
	DB: config.DB,
}

func init() {
	UserRepo.DB.AutoMigrate(&model.User{})
}
