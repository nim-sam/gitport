package commithistory

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/nim-sam/gitport/internal/model"
)

// Commit Selection Message
type CommitSelectionMsg struct {
	commit model.Commit
}

// CommitSelectionMsg commit getter
func (msg CommitSelectionMsg) GetCommit() model.Commit {
	return msg.commit
}

// Create CommitSelectionMsg containing currently selected commit
func newCommitSelectionMsg(commit model.Commit) tea.Cmd {
	return func() tea.Msg {
		return CommitSelectionMsg{
			commit: commit,
		}
	}
}
