package auth

import (
	"github.com/nim-sam/gitport/internal/config"
	"github.com/nim-sam/gitport/pkg/repository"
	"gorm.io/gorm"
)

// Enum categorizing access levels into three levels : Read Only, Read & Write, Admin
type AccessLevel int

const (
	ReadOnly AccessLevel = iota
	ReadWrite
	Admin
)

// Struct holding user information
type User struct {
	gorm.Model
	SshKey      string
	Name        string
	Email       string
	AccessLevel AccessLevel
}

// User Repository
var UserRepo = repository.Repository[User]{
	DB: config.DB,
}

func init() {
	UserRepo.DB.AutoMigrate(&User{})
}
