package commithistory

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nim-sam/gitport/internal/model"
	"github.com/nim-sam/gitport/ui"
)

/*
 * Commit history is a bubbles list wrapper used to visualize the repo's most recent
 * commits. It is located in the left panel of the commit history dashboard tab.
 * It also provides helper methods to get the currently selected commit from
 * the list model . It also passes messages to the commit viewport, such that it can
 * update its content according to the currently selected commit item.
 */

// Commit History Model
type Model struct {
	selectedCommit int
	ListModel      list.Model
	commits        []model.Commit
	Width          int
	Height         int
	focused        bool
}

func (c Model) Init() tea.Cmd {
	// Sends currently selected commit instance through a CommitSelectionMsg upon
	// model initialization
	return newCommitSelectionMsg(
		c.GetCurrentCommit(),
	)

}

func (c Model) Update(msg tea.Msg) (Model, tea.Cmd) {

	var cmd tea.Cmd

	// Update list model
	c.ListModel.SetSize(c.Width, c.Height)

	if c.focused {
		c.ListModel, _ = c.ListModel.Update(msg)

		// Send currently selected commit instance through a CommitSelectionMsg upon
		// updating the list model
		cmd = newCommitSelectionMsg(
			c.GetCurrentCommit(),
		)
	}

	return c, cmd
}
func (c Model) View() string {

	container := lipgloss.NewStyle().
		Width(c.Width)

	return container.Render(c.ListModel.View())
}

func New(commits []model.Commit) Model {

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

	return Model{
		commits:   commits,
		ListModel: model,
	}
}

// Sets m.focused value
func (m *Model) SetFocused(value bool) {
	m.focused = value
}

// Returns currenty selected commit
func (m Model) GetCurrentCommit() model.Commit {
	index := m.ListModel.Index()
	return m.commits[index]
}
