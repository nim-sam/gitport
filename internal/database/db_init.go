package database

import (
	"os"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Init DB
var DB *gorm.DB

// Creates .gitport folder if not present in the repo and opens an SQLite database
func init() {
	os.MkdirAll("./.gitport", 0755)
	out, err := gorm.Open(sqlite.Open("./.gitport/gitport.db"), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	DB = out
}
