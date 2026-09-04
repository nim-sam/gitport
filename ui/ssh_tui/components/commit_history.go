package components

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nim-sam/gitport/internal/model"
	"github.com/nim-sam/gitport/ui"
)

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

type CommitHistoryModel struct {
	selectedCommit int
	ListModel      list.Model
	commits        []model.Commit
	Width          int
	Height         int
}

func (c CommitHistoryModel) Init() tea.Cmd { return nil }
func (c CommitHistoryModel) Update(msg tea.Msg) (CommitHistoryModel, tea.Cmd) {
	c.ListModel.SetSize(c.Width, c.Height)
	c.ListModel, _ = c.ListModel.Update(msg)
	return c, nil
}
func (c CommitHistoryModel) View() string {

	container := lipgloss.NewStyle().
		Width(c.Width)
		// Padding(0, 1)

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
	delegate.Styles.NormalTitle = delegate.Styles.NormalTitle.
		Foreground(ui.Palette.FgHighlight)
	delegate.Styles.NormalDesc = delegate.Styles.NormalDesc.
		Italic(true).
		Foreground(ui.Palette.FgMuted)
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

// Returns currenty selected commit
func (m CommitHistoryModel) GetCurrentCommit() model.Commit {
	index := m.ListModel.Index()
	return m.commits[index]
}
