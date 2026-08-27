package git

import (
	"github.com/nim-sam/gitport/internal/config"
	"github.com/nim-sam/gitport/pkg/repository"
	"gorm.io/gorm"
)

// Struct holding commit information
type Commit struct {
	gorm.Model
	Author  string
	Hash    string
	Message string
	Date    string
	Branch  string
}

// Commit Repository
var CommitRepo = repository.Repository[Commit]{
	DB: config.DB,
}

func init() {
	CommitRepo.DB.AutoMigrate(&Commit{})
}
