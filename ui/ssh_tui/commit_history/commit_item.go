package commithistory

import "strings"

// CommitItem implements list.Item interface
type CommitItem struct {
	author  string
	hash    string
	message string
	date    string
}

func (i CommitItem) FilterValue() string {
	return i.hash
}

func (i CommitItem) Title() string {
	return "[" + i.hash + "] " + i.subject()
}

// subject is the first line of the commit message. go-git stores the full
// message (subject plus optional body) on one string, and a body would
// otherwise render as extra lines in the list.
func (i CommitItem) subject() string {
	subject, _, _ := strings.Cut(i.message, "\n")
	return strings.TrimSpace(subject)
}

func (i CommitItem) Description() string {
	return i.author + " @ " + i.date
}
