package auth

// Enum categorizing access levels into three levels : Read Only, Read & Write, Admin
type AccessLevel int

const (
	ReadOnly AccessLevel = iota
	ReadWrite
	Admin
)

// Struct holding user information
type User struct {
	sshKey      string
	name        string
	accessLevel AccessLevel
}

// User constructor
func NewUser() *User {
	return &User{}
}

// User getters
func (u *User) SshKey() string           { return u.sshKey }
func (u *User) Name() string             { return u.name }
func (u *User) AccessLevel() AccessLevel { return u.accessLevel }

// User setters
func (u *User) WithSshKey(key string) *User                   { u.sshKey = key; return u }
func (u *User) WithName(name string) *User                    { u.name = name; return u }
func (u *User) WithAccessLevel(accessLevel AccessLevel) *User { u.accessLevel = accessLevel; return u }
