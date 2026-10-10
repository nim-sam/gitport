package initprompter

import (
	"github.com/nim-sam/gitport/internal/config"
	"github.com/nim-sam/gitport/internal/model"
)

// InitConfigMsg carries the configuration gathered by the init prompter back
// to the command handler, which persists it.
type InitConfigMsg struct {
	Config config.Config
	Admin  model.User
}

func (m InitConfigMsg) GetConfig() config.Config {
	return m.Config
}

func (m InitConfigMsg) GetAdmin() model.User {
	return m.Admin
}
