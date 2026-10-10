// Package sshkey validates and compares the OpenSSH public keys stored on user
// accounts. Keys are held in authorized_keys form, i.e. the "<type> <blob>"
// line optionally followed by a comment.
package sshkey

import (
	"strings"

	"github.com/charmbracelet/ssh"
)

// Parse validates an authorized_keys-formatted public key. Leading and
// trailing whitespace is trimmed so values pasted with a trailing newline
// still parse.
func Parse(raw string) (ssh.PublicKey, error) {
	key, _, _, _, err := ssh.ParseAuthorizedKey([]byte(strings.TrimSpace(raw)))
	return key, err
}

// Matches reports whether key equals the public key encoded in raw. A blank or
// malformed raw never matches.
func Matches(raw string, key ssh.PublicKey) bool {
	stored, err := Parse(raw)
	if err != nil {
		return false
	}
	return ssh.KeysEqual(key, stored)
}
