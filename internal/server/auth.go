package server

import (
	"github.com/charmbracelet/ssh"

	"github.com/nim-sam/gitport/internal/repository"
	"github.com/nim-sam/gitport/internal/sshkey"
)

// authorize accepts a connection only when the presented public key matches one
// stored on a user in the database. Users are reloaded on every attempt so
// accounts added through the dashboard take effect without a restart.
func (s Server) authorize(_ ssh.Context, key ssh.PublicKey) bool {
	for _, u := range repository.UserRepo.Load() {
		if sshkey.Matches(u.SshKey, key) {
			return true
		}
	}
	return false
}
