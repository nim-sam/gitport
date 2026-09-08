package commitviewport

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nim-sam/gitport/internal/config"
	"github.com/nim-sam/gitport/ui"
	commithistory "github.com/nim-sam/gitport/ui/ssh_tui/commit_history"
)

/*
 * The commit viewport is the component found in the right panel of the commit history
 * tab of the GitPort dashboard. The commit viewport displays the commit diff of the
 * currently selected commit by the commit history model. The reason as to why a plain
 * viewport wasn't used was for better message passing management and decluttering
 * the dashboard code from lipgloss styling
 */

// Viewport wrapper to visualize git diffs
type Model struct {
	viewport viewport.Model
	width    int
	height   int
	focused  bool
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {

	// Set dimensions
	m.viewport.Height = m.height
	m.viewport.Width = m.width

	switch msg := msg.(type) {

	// Recieves CommitSelectionMsg from commit history model and sets
	// viewport content with commit diff based on the commit contained
	// in the message
	case commithistory.CommitSelectionMsg:
		commit := msg.GetCommit()
		diff := config.GitportServer.GetCommitDiff(commit.Hash)
		m.SetContent(diff)
		m.viewport.SetYOffset(0)
	}

	if m.focused {
		m.viewport, _ = m.viewport.Update(msg)
	}

	return m, nil
}

func (m Model) View() string { return m.viewport.View() }

// Width and height setters
func (m *Model) Width(w int) *Model  { m.width = w; return m }
func (m *Model) Height(h int) *Model { m.height = h; return m }

// Styles git diff string and sets viewport's content
func (m *Model) SetContent(diff string) {

	textStyle := lipgloss.NewStyle().
		Foreground(ui.Palette.Fg)

	m.viewport.SetContent(textStyle.Render(diff))
}

// Sets m.focused value
func (m *Model) SetFocused(value bool) { m.focused = value }

// Create new commit viewport
func New() Model {
	return Model{
		viewport: viewport.New(0, 0),
		height:   0,
		width:    0,
	}
}
