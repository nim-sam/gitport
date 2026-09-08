package components

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nim-sam/gitport/internal/config"
	"github.com/nim-sam/gitport/ui"
)

// Custom viewport wrapper to visualize git diffs
type CommitViewport struct {
	viewport viewport.Model
	width    int
	height   int
	focused  bool
}

func (m CommitViewport) Init() tea.Cmd { return nil }

func (m CommitViewport) Update(msg tea.Msg) (CommitViewport, tea.Cmd) {

	// Set dimensions
	m.viewport.Height = m.height
	m.viewport.Width = m.width

	switch msg := msg.(type) {

	// Recieves CommitSelectionMsg from commit history model and sets
	// viewport content with commit diff based on the commit contained
	// in the message
	case CommitSelectionMsg:
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

func (m CommitViewport) View() string { return m.viewport.View() }

// Width and height setters
func (m *CommitViewport) Width(w int) *CommitViewport  { m.width = w; return m }
func (m *CommitViewport) Height(h int) *CommitViewport { m.height = h; return m }

// Styles git diff string and sets viewport's content
func (m *CommitViewport) SetContent(diff string) {

	textStyle := lipgloss.NewStyle().
		Foreground(ui.Palette.Fg)

	m.viewport.SetContent(textStyle.Render(diff))
}

// Sets m.focused value
func (m *CommitViewport) SetFocused(value bool) { m.focused = value }

// Create new commit viewport
func NewCommitViewport() CommitViewport {
	return CommitViewport{
		viewport: viewport.New(0, 0),
		height:   0,
		width:    0,
	}
}
