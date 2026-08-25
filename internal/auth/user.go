package auth

import "gorm.io/gorm"

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
