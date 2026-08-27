package model

import (
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
