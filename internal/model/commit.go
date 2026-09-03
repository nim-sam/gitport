package model

import (
	_ "github.com/charmbracelet/wish/git"
	"gorm.io/gorm"
)

// Struct holding commit information
type Commit struct {
	gorm.Model
	Author  string
	Hash    string
	Message string
	Date    string
}
