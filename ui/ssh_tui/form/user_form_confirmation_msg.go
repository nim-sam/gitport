package form

import "github.com/nim-sam/gitport/internal/model"

// User form confirmation msg
type UserFormConfirmationMsg struct {
	user model.User
}

func (u UserFormConfirmationMsg) GetUser() model.User {
	return u.user
}
