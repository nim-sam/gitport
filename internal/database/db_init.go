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
	// 0700/0600 because the database stores user SSH public keys and emails.
	// Existing paths are tightened too, since earlier versions created them 0755/0644.
	if err := os.MkdirAll("./.gitport", 0700); err != nil {
		panic(err)
	}
	if err := os.Chmod("./.gitport", 0700); err != nil {
		panic(err)
	}

	out, err := gorm.Open(sqlite.Open("./.gitport/gitport.db"), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	if err := os.Chmod("./.gitport/gitport.db", 0600); err != nil {
		panic(err)
	}
	DB = out
}
