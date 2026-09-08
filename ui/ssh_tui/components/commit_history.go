package components

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nim-sam/gitport/internal/model"
	"github.com/nim-sam/gitport/ui"
)

// CommitItem implements list.Item interface
type CommitItem struct {
	author  string
	hash    string
	message string
	date    string
}

func (i CommitItem) FilterValue() string { return i.hash }
func (i CommitItem) Title() string {
	return "[" + i.hash + "] " + i.message
}
func (i CommitItem) Description() string {
	return i.author + " @ " + i.date
}

// Commit History Model
type CommitHistoryModel struct {
	selectedCommit int
	ListModel      list.Model
	commits        []model.Commit
	Width          int
	Height         int
	focused        bool
}

// Commit Selection Message
type CommitSelectionMsg struct {
	commit model.Commit
}

func (c CommitHistoryModel) Init() tea.Cmd {
	// Sends currently selected commit instance through a CommitSelectionMsg upon
	// model initialization
	return createCommitSelectionMsg(
		c.GetCurrentCommit(),
	)

}

func (c CommitHistoryModel) Update(msg tea.Msg) (CommitHistoryModel, tea.Cmd) {

	var cmd tea.Cmd

	// Update list model
	c.ListModel.SetSize(c.Width, c.Height)

	if c.focused {
		c.ListModel, _ = c.ListModel.Update(msg)

		// Send currently selected commit instance through a CommitSelectionMsg upon
		// updating the list model
		cmd = createCommitSelectionMsg(
			c.GetCurrentCommit(),
		)
	}

	return c, cmd
}
func (c CommitHistoryModel) View() string {

	container := lipgloss.NewStyle().
		Width(c.Width)

	return container.Render(c.ListModel.View())
}

func NewCommitHistory(commits []model.Commit) CommitHistoryModel {

	// Initializing commit items
	var commitItems []list.Item
	for _, commit := range commits {
		commitItems = append(commitItems,
			CommitItem{
				author:  commit.Author,
				hash:    commit.Hash[:6],
				message: commit.Message,
				date:    commit.Date,
			},
		)
	}

	// Defining new delegate
	delegate := list.NewDefaultDelegate()

	// Selected
	delegate.Styles.SelectedTitle = lipgloss.NewStyle().
		Foreground(ui.Palette.Accent).
		Border(lipgloss.HiddenBorder(), false, false, false, true).
		Padding(0, 1).
		BorderForeground(ui.Palette.Accent).
		Bold(true)
	delegate.Styles.SelectedDesc = lipgloss.NewStyle().
		Foreground(ui.Palette.Accent).
		Border(lipgloss.HiddenBorder(), false, false, false, true).
		Padding(0, 1).
		BorderForeground(ui.Palette.Accent).
		Italic(true)

	// Normal
	delegate.Styles.NormalTitle = delegate.Styles.NormalTitle.
		Foreground(ui.Palette.FgHighlight)
	delegate.Styles.NormalDesc = delegate.Styles.NormalDesc.
		Italic(true).
		Foreground(ui.Palette.FgMuted)

	// Dimmed
	delegate.Styles.DimmedTitle = lipgloss.NewStyle().
		Foreground(ui.Palette.Fg).
		Border(lipgloss.HiddenBorder(), false, false, false, true).
		Padding(0, 1).
		Faint(true)
	delegate.Styles.DimmedDesc = lipgloss.NewStyle().
		Foreground(ui.Palette.FgMuted).
		Border(lipgloss.HiddenBorder(), false, false, false, true).
		Padding(0, 1).
		Italic(true).
		Faint(true)

	// Creating list model
	model := list.New(commitItems, delegate, 0, 0)
	model.SetShowHelp(false)
	model.SetShowTitle(false)
	model.SetShowStatusBar(false)

	return CommitHistoryModel{
		commits:   commits,
		ListModel: model,
	}
}

// Sets m.focused value
func (m *CommitHistoryModel) SetFocused(value bool) {
	m.focused = value
}

// Returns currenty selected commit
func (m CommitHistoryModel) GetCurrentCommit() model.Commit {
	index := m.ListModel.Index()
	return m.commits[index]
}

// CommitSelectionMsg commit getter
func (msg CommitSelectionMsg) GetCommit() model.Commit {
	return msg.commit
}

// Create CommitSelectionMsg containing currently selected commit
func createCommitSelectionMsg(commit model.Commit) tea.Cmd {
	return func() tea.Msg {
		return CommitSelectionMsg{
			commit: commit,
		}
	}
}
