package components

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nim-sam/gitport/ui"
)

// Custom viewport wrapper to visualize git diffs
type CommitViewport struct {
	viewport viewport.Model
	width    int
	height   int
}

func (m CommitViewport) Init() tea.Cmd { return nil }
func (m CommitViewport) Update(msg tea.Msg) (CommitViewport, tea.Cmd) {
	m.viewport.Height = m.height
	m.viewport.Width = m.width
	m.viewport, _ = m.viewport.Update(msg)
	return m, nil
}

func (m CommitViewport) View() string {
	return m.viewport.View()
}

func (m *CommitViewport) Width(w int) *CommitViewport  { m.width = w; return m }
func (m *CommitViewport) Height(h int) *CommitViewport { m.height = h; return m }

// Styles git diff string and sets viewport's content
func (m *CommitViewport) SetContent(diff string) {

	textStyle := lipgloss.NewStyle().
		Foreground(ui.Palette.Fg)

	m.viewport.SetContent(textStyle.Render(diff))
}

// Create new commit viewport
func NewCommitViewport() CommitViewport {
	return CommitViewport{
		viewport: viewport.New(0, 0),
		height:   0,
		width:    0,
	}
}
