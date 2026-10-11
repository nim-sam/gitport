package server

import (
	"github.com/charmbracelet/ssh"

	"github.com/nim-sam/gitport/internal/model"
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

// authenticatedUser resolves the database user whose key authorized the
// session. It returns false when the session carries no public key or the key
// no longer belongs to any user, in which case callers should fall back to the
// least-privileged role.
func (s Server) authenticatedUser(sess ssh.Session) (model.User, bool) {
	key := sess.PublicKey()
	if key == nil {
		return model.User{}, false
	}
	for _, u := range repository.UserRepo.Load() {
		if sshkey.Matches(u.SshKey, key) {
			return u, true
		}
	}
	return model.User{}, false
}
