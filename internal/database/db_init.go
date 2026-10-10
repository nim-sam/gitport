package database

import (
	"fmt"
	"os"
	"path/filepath"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB is the process-wide database handle. It is nil until Init runs, so any
// caller must initialize the database before touching the repositories.
var DB *gorm.DB

// Init creates the .gitport folder under root (if absent) and opens the SQLite
// database stored inside it, leaving the handle in DB.
//
// Permission bits are 0700/0600 because the database stores user SSH public
// keys and emails. Existing paths are tightened too, since earlier versions
// created them 0755/0644.
func Init(root string) error {
	dir := filepath.Join(root, ".gitport")

	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("could not create %s: %w", dir, err)
	}
	if err := os.Chmod(dir, 0700); err != nil {
		return fmt.Errorf("could not secure %s: %w", dir, err)
	}

	dbPath := filepath.Join(dir, "gitport.db")

	// Silent GORM's logger: it writes to stdout by default and would dump
	// slow queries (with the full commit diff) over the Bubble Tea TUI.
	out, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return fmt.Errorf("could not open database: %w", err)
	}
	if err := os.Chmod(dbPath, 0600); err != nil {
		return fmt.Errorf("could not secure %s: %w", dbPath, err)
	}

	DB = out
	return nil
}
