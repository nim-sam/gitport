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

// CanClone reports whether the level is allowed to read the served repository
// by cloning. Every defined level may; it exists so the dashboard permission
// matrix and the git access mapping share one source of truth.
func (a AccessLevel) CanClone() bool {
	return a >= ReadOnly
}

// CanPull reports whether the level is allowed to fetch updates. Cloning and
// pulling travel over the same git-upload-pack command, so they share a gate.
func (a AccessLevel) CanPull() bool {
	return a >= ReadOnly
}

// CanPush reports whether the level is allowed to write to the served
// repository. ReadWrite and above may; ReadOnly may not.
func (a AccessLevel) CanPush() bool {
	return a >= ReadWrite
}

// Struct holding user information
type User struct {
	gorm.Model
	SshKey      string
	Name        string
	Email       string
	AccessLevel AccessLevel
}
