package auth

type accessLevel int

const (
	ReadOnly accessLevel = iota
	ReadWrite
)

type User struct {
	sshKey string
	name   string
	access accessLevel
}

func (u User) SshKey() string {
	return u.sshKey
}

func (u User) Name() string {
	return u.name
}

func (u User) AccessLevel() string {
	return u.accessLevel
}
