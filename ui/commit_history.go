package ui

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nim-sam/gitport/internal/git"
)

type CommitItem struct {
	author  string
	hash    string
	message string
	date    string
	branch  string
}

func (i CommitItem) FilterValue() string { return i.hash }
func (i CommitItem) Title() string {
	return "[" + i.hash + "] " + i.message
}
func (i CommitItem) Description() string {
	return i.author + " @ " + i.date + " (" + i.branch + ")"
}

type CommitHistoryModel struct {
	selectedCommit int
	listModel      list.Model
	commits        []*git.Commit
	width          int
	heigth         int
}

func (c CommitHistoryModel) Init() tea.Cmd { return nil }
func (c CommitHistoryModel) Update(msg tea.Msg) (CommitHistoryModel, tea.Cmd) {
	c.listModel.SetSize(c.width, c.heigth)
	c.listModel, _ = c.listModel.Update(msg)
	return c, nil
}
func (c CommitHistoryModel) View() string {

	container := lipgloss.NewStyle().
		Width(c.width)
		// Padding(0, 1)

	return container.Render(c.listModel.View())
}

func NewCommitHistory(commits []*git.Commit) CommitHistoryModel {

	// Initializing commit items
	var commitItems []list.Item
	for _, commit := range commits {
		commitItems = append(commitItems,
			CommitItem{
				author:  commit.Author(),
				hash:    commit.Hash(),
				message: commit.Message(),
				date:    commit.Date(),
				branch:  commit.Branch(),
			},
		)
	}

	// Defining new delegate
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = lipgloss.NewStyle().
		Foreground(palette.accent).
		Border(lipgloss.HiddenBorder(), false, false, false, true).
		Padding(0, 1).
		BorderForeground(palette.accent).
		Bold(true)
	delegate.Styles.SelectedDesc = lipgloss.NewStyle().
		Foreground(palette.accent).
		Border(lipgloss.HiddenBorder(), false, false, false, true).
		Padding(0, 1).
		BorderForeground(palette.accent).
		Italic(true)
	delegate.Styles.NormalTitle = delegate.Styles.NormalTitle.
		Foreground(palette.fgHighlight)
	delegate.Styles.NormalDesc = delegate.Styles.NormalDesc.
		Italic(true).
		Foreground(palette.fgMuted)

	// Creating list model
	model := list.New(commitItems, delegate, 0, 0)
	model.SetShowHelp(false)
	model.SetShowTitle(false)
	model.SetShowStatusBar(false)

	return CommitHistoryModel{
		commits:   commits,
		listModel: model,
	}
}
