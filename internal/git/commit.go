package git

// Struct holding commit information
type Commit struct {
	author  string
	hash    string
	message string
	date    string
	branch  string
}

// Commit constructor
func NewCommit() *Commit {
	return &Commit{}
}

// Commit getters
func (c *Commit) Author() string  { return c.author }
func (c *Commit) Hash() string    { return c.hash }
func (c *Commit) Message() string { return c.message }
func (c *Commit) Date() string    { return c.date }
func (c *Commit) Branch() string  { return c.branch }

// Commit setters
func (c *Commit) WithAuthor(name string) *Commit   { c.author = name; return c }
func (c *Commit) WithHash(hash string) *Commit     { c.hash = hash; return c }
func (c *Commit) WithMessage(msg string) *Commit   { c.message = msg; return c }
func (c *Commit) WithDate(date string) *Commit     { c.date = date; return c }
func (c *Commit) WithBranch(branch string) *Commit { c.branch = branch; return c }
