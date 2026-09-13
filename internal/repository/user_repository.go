package repository

import (
	"github.com/nim-sam/gitport/internal/database"
	"github.com/nim-sam/gitport/internal/model"
	orm "github.com/nim-sam/gitport/pkg/mini_orm"
)

// User Repository
var UserRepo = orm.Repository[model.User]{
	DB: database.DB,
}

func init() {
	UserRepo.DB.AutoMigrate(&model.User{})
}
