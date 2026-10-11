package model

import (
	"gorm.io/gorm"
)

// Enum categorizing access levels : Read Only, Read & Write, Admin, Owner.
// The order matters: the dashboard form's role selector maps its item index
// straight to these values.
type AccessLevel int

const (
	ReadOnly AccessLevel = iota
	ReadWrite
	Admin
	Owner
)

// CanManage reports whether the level is allowed to manage users through the
// dashboard. Admins and Owners can; ReadOnly and ReadWrite cannot.
func (a AccessLevel) CanManage() bool {
	return a >= Admin
}

// Struct holding user information
type User struct {
	gorm.Model
	SshKey      string
	Name        string
	Email       string
	AccessLevel AccessLevel
}
