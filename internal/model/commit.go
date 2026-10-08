package model

import (
	"gorm.io/gorm"
)

// DateLayout is the layout Commit.Date is formatted with when loading from
// git, and the layout it must be parsed with when ordering commits by date.
const DateLayout = "Jan 02 2006"

// Struct holding commit information
type Commit struct {
	gorm.Model
	Author  string
	Hash    string
	Message string
	Date    string
	Diff    string
}
