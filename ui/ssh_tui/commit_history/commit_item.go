package commithistory

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
	return "[" + i.hash + "] " + i.message
}

func (i CommitItem) Description() string {
	return i.author + " @ " + i.date
}
