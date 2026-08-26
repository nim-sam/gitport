package logs

import "gorm.io/gorm"

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
