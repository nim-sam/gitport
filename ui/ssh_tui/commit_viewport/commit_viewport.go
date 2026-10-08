package commitviewport

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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

// Styles

var (
	normalText  = lipgloss.NewStyle().Foreground(ui.Palette.FgMuted)
	addedLine   = lipgloss.NewStyle().Foreground(ui.Palette.FgHighlight).Bold(true)
	removedLine = lipgloss.NewStyle().Foreground(ui.Palette.Fg).Italic(true)
)

// Apply styles to diff text
func stylizeDiff(diff string) string {

	diffLines := strings.Split(diff, "\n")
	var out []string

	for _, line := range diffLines {
		// Skip if line is empty
		if len(line) == 0 {
			continue
		}
		switch line[0] {
		case '+':
			line = addedLine.Render(line)
		case '-':
			line = removedLine.Render(line)
		default:
			line = normalText.Render(line)
		}

		out = append(out, line)
	}

	return strings.Join(out, "\n")
}

// Viewport wrapper to visualize git diffs
type Model struct {
	viewport viewport.Model
	width    int
	height   int
	focused  bool

	// Takes care of keeping the offset of each viewport from
	// resetting each time a CommitSelectionMsg is emitted
	offset             map[string]int
	selectedCommitHash string

	// Display data
	commitHashShort string
	commitMsg       string
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {

	// Set dimensions
	m.viewport.Height = m.height - 2
	m.viewport.Width = m.width - 2

	switch msg := msg.(type) {

	// Recieves CommitSelectionMsg from commit history model and sets
	// viewport content with commit diff based on the commit contained
	// in the message
	case commithistory.CommitSelectionMsg:

		commit := msg.GetCommit()
		m.selectedCommitHash = commit.Hash

		// The diff travels on the commit itself, so no lookup is needed
		diff := commit.Diff
		if diff == "" {
			diff = "No diff available."
		}
		m.SetContent(diff)

		m.viewport.SetYOffset(m.offset[m.selectedCommitHash])

		// Setting commit display values
		m.commitHashShort = shortHash(commit.Hash)
		m.commitMsg = commit.Message
	}

	if m.focused {
		m.viewport, _ = m.viewport.Update(msg)
		m.offset[m.selectedCommitHash] = m.viewport.YOffset
	}

	return m, nil
}

func (m Model) View() string {

	diffContainer := lipgloss.NewStyle().
		Height(m.height-2).
		Width(m.width).
		Padding(0, 1)

	diffTitle := "[" + m.commitHashShort + "] " + m.commitMsg

	if len(diffTitle) > m.width {
		diffTitle = diffTitle[:m.width-5]
		diffTitle += "…"
	}

	diffTitleRender := lipgloss.NewStyle().
		Width(m.width-2).
		Height(1).
		Align(lipgloss.Center).
		Bold(true).
		Foreground(ui.Palette.Fg).
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(ui.Palette.Border).
		Render(diffTitle)

	titleCentered := lipgloss.NewStyle().
		AlignHorizontal(lipgloss.Center).
		Width(m.width).
		Render(diffTitleRender)

	return lipgloss.JoinVertical(
		lipgloss.Top,
		titleCentered,
		diffContainer.Render(m.viewport.View()),
	)
}

// Width and height setters
func (m *Model) Width(w int) *Model  { m.width = w; return m }
func (m *Model) Height(h int) *Model { m.height = h; return m }

// shortHash returns the abbreviated hash shown in the viewport title. The
// zero commit of an empty repository has an empty hash, so shorter inputs are
// returned unchanged instead of panicking on a slice bound.
func shortHash(hash string) string {
	if len(hash) > 6 {
		return hash[:6]
	}
	return hash
}

// Styles git diff string and sets viewport's content
func (m *Model) SetContent(diff string) {
	m.viewport.SetContent(stylizeDiff(diff))
}

// Sets m.focused value
func (m *Model) SetFocused(value bool) { m.focused = value }

// Create new commit viewport
func New() Model {
	return Model{
		viewport: viewport.New(0, 0),
		height:   0,
		width:    0,
		offset:   make(map[string]int),
	}
}
