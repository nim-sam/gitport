package logs

import (
	"github.com/nim-sam/gitport/internal/config"
	"github.com/nim-sam/gitport/pkg/repository"
	"gorm.io/gorm"
)

// Enum categorizing logs into three levels : Warning, Error, Info
type LogLevel int

const (
	Warning LogLevel = iota
	Error
	Info
)

// Struct holding server log information
type Log struct {
	gorm.Model
	Level   LogLevel
	Date    string
	Message string
}

// Logs Repository
var LogRepo = repository.Repository[Log]{
	DB: config.DB,
}

func init() {
	LogRepo.DB.AutoMigrate(&Log{})
}
