package userlist

import (
	"fmt"

	"github.com/nim-sam/gitport/internal/model"
)

// User Item definition

type userItem struct {
	id          uint
	sshKey      string
	name        string
	email       string
	accessLevel model.AccessLevel
}

// User's list.Item implementation

func (i userItem) FilterValue() string { return i.name }
func (i userItem) Title() string       { return i.name + " <" + i.email + ">" }
func (i userItem) Description() string {

	var accessLevelDisplay string
	switch i.accessLevel {
	case model.ReadOnly:
		accessLevelDisplay = "(Read)"
	case model.ReadWrite:
		accessLevelDisplay = "(Write)"
	case model.Admin:
		accessLevelDisplay = "(Admin)"
	case model.Owner:
		accessLevelDisplay = "(Owner)"
	}

	return fmt.Sprintf("%s %s", accessLevelDisplay, i.sshKey)
}
