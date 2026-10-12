package userdetail

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	lipglosstable "github.com/charmbracelet/lipgloss/table"
	"github.com/nim-sam/gitport/internal/model"
	"github.com/nim-sam/gitport/ui"
)

/*
 * The user detail component is the right panel of the user management tab. It
 * owns the panel's rounded border and lays its content out as a two row table:
 * the top row holds the selected user's cumulative activity, the bottom row
 * their most recent commits. It holds no database handle: the dashboard feeds
 * it the selected user, their activity record and their commits, so a render
 * never scans the log table.
 */

// commitLimit caps how many of the selected user's commits are listed.
const commitLimit = 15

// Fixed column widths for the commit table; the message column absorbs the
// remaining width. The bubbles table pads every cell by one space on each side,
// which the message width accounts for.
const (
	hashColWidth = 7
	dateColWidth = 11
	cellPadding  = 2
)

// Vertical budget of the panel. The activity row is the username plus one line
// per tracked action; chromeLines covers the top border, the divider between
// the two rows and the bottom border. Everything left over is commit table.
const (
	activityLines = 1 + 3
	chromeLines   = 3
)

var (
	labelStyle = lipgloss.NewStyle().Foreground(ui.Palette.FgMuted)
	valueStyle = lipgloss.NewStyle().Foreground(ui.Palette.Fg)
	mutedStyle = lipgloss.NewStyle().Foreground(ui.Palette.FgMuted)
	nameStyle  = lipgloss.NewStyle().Bold(true).Foreground(ui.Palette.Fg)
)

type Model struct {
	Width            int
	Height           int
	user             model.User
	activity         model.Activity
	table            table.Model
	focused          bool
	mostRecentCommit model.Commit
}

func (m Model) Init() tea.Cmd { return nil }

// Update forwards key events to the commit table while the panel is focused so
// the table's viewport can be scrolled.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.focused {
		return m, nil
	}
	var cmd tea.Cmd
	// m.table, cmd = m.table.Update(msg)
	return m, cmd
}

// SetUser sets the user whose name titles the panel.
func (m *Model) SetUser(u model.User) { m.user = u }

// SetActivity sets the cumulative activity shown in the top row.
func (m *Model) SetActivity(a model.Activity) { m.activity = a }

// SetFocused reports whether the panel currently owns keyboard focus, which
// selects the border colour and whether the commit table handles scroll keys.
func (m *Model) SetFocused(f bool) {
	m.focused = f
	if f {
		m.table.Focus()
	} else {
		m.table.Blur()
	}
}

// SetCommits replaces the commit table with commits authored by the selected
// user, newest first. At most commitLimit rows are shown.
func (m *Model) SetCommits(commits []model.Commit) {

	if len(commits) > commitLimit {
		commits = commits[:commitLimit]
	}

	if len(commits) > 0 {
		m.mostRecentCommit = commits[0]
		commits = commits[1:]
	} else {
		m.mostRecentCommit = model.Commit{}
	}

	rows := make([]table.Row, 0, len(commits))
	for _, c := range commits {
		rows = append(rows, table.Row{
			shortHash(c.Hash),
			c.Date,
			subject(c.Message),
		})
	}

	m.table.SetRows(rows)
	m.table.SetCursor(0)
	m.table.GotoTop()
}

// SetSize gives the panel its total footprint, border included.
func (m *Model) SetSize(width, height int) {
	m.Width = width
	m.Height = height

	inner := m.innerWidth()
	cols := m.table.Columns()
	if len(cols) == 3 {
		cols[0].Width = hashColWidth
		cols[1].Width = dateColWidth
		cols[2].Width = inner - hashColWidth - dateColWidth - 3*cellPadding
		if cols[2].Width < 6 {
			cols[2].Width = 6
		}
		m.table.SetColumns(cols)
	}

	m.table.SetWidth(inner)
	m.table.SetHeight(m.Height - 11)
}

func (m Model) View() string {

	// sectionTitleStyle := lipgloss.NewStyle().
	// 	Width(m.Width - 2).
	// 	AlignHorizontal(lipgloss.Center).
	// 	Italic(true).
	// 	Foreground(ui.Palette.FgMuted)

	nameTitleStyle := lipgloss.NewStyle().
		Width(m.Width - 2).
		AlignHorizontal(lipgloss.Center).
		Bold(true).
		Foreground(ui.Palette.FgHighlight)

	mostRecentCommitTitle := lipgloss.NewStyle().
		Foreground(ui.Palette.FgHighlight).
		Bold(true)

	mostRecentCommitMetadata := lipgloss.NewStyle().
		Foreground(ui.Palette.FgMuted).
		Italic(true)

	mostRecentCommitContainer := lipgloss.NewStyle().Height(2).Width(m.Width-4).Padding(0, 1)

	var mostRecentCommitView string
	if (m.mostRecentCommit != model.Commit{}) {

		commitTitle := "[" + m.mostRecentCommit.Hash[:6] + "] " + subject(m.mostRecentCommit.Message)
		if len(commitTitle) > m.Width-7 {
			commitTitle = commitTitle[:m.Width-7] + "…"
		}

		mostRecentCommitView = mostRecentCommitContainer.Render(
			lipgloss.JoinVertical(
				lipgloss.Top,
				mostRecentCommitTitle.Render(
					commitTitle,
				),
				mostRecentCommitMetadata.Render(
					m.mostRecentCommit.Author+" @ "+m.mostRecentCommit.Date,
				),
			),
		)
	} else {
		mostRecentCommitView = mostRecentCommitContainer.Foreground(ui.Palette.FgMuted).Render(
			"No available commit data.",
		)
	}

	tableContainerStyle := lipgloss.NewStyle().Foreground(ui.Palette.FgMuted)

	contentContainer := lipglosstable.New().
		Border(lipgloss.RoundedBorder()).
		BorderRow(true).
		BorderStyle(lipgloss.NewStyle().Foreground(ui.Palette.Border)).
		Width(m.Width).
		Height(m.Height).
		Rows(
			[]string{
				nameTitleStyle.Render(m.user.Name),
			},
			// []string{
			// 	sectionTitleStyle.Render("Activity"),
			// },
			[]string{
				m.activityRow(),
			},
			// []string{
			// 	sectionTitleStyle.Render("Latest Commit"),
			// },
			[]string{
				mostRecentCommitView,
			},
			// []string{
			// 	sectionTitleStyle.Render("History"),
			// },
			[]string{
				tableContainerStyle.Render(m.table.View()),
				// m.table.View(),
			},
		)

	if m.focused {
		contentContainer = contentContainer.
			BorderStyle(
				lipgloss.NewStyle().
					Foreground(ui.Palette.SelectedBorder),
			)
	}

	return contentContainer.Render()
}

// activityRow renders the panel's top row: the user's name followed by their
// last login, push and fetch alongside running totals. Times with no recorded
// event read "never".
func (m Model) activityRow() string {
	rows := []struct {
		label string
		when  time.Time
		count int
	}{
		{"◈ login", m.activity.LastLogin, m.activity.Logins},
		{"⤒ push", m.activity.LastPush, m.activity.Pushes},
		{"⤓ fetch", m.activity.LastFetch, m.activity.Fetches},
	}

	lines := []string{}
	for _, r := range rows {
		label := labelStyle.Width(10).Render("  " + r.label)
		when := valueStyle.Width(12).Render(relative(r.when))
		count := mutedStyle.Render(fmt.Sprintf("%d total", r.count))
		lines = append(lines, label+when+count)
	}

	return lipgloss.JoinVertical(
		lipgloss.Top,
		lines...,
	)
}

// innerWidth is the number of cells available between the panel's side borders.
func (m Model) innerWidth() int {
	if m.Width > 2 {
		return m.Width - 2
	}
	return m.Width
}

// shortHash abbreviates a commit hash. The zero commit of an empty repository
// has an empty hash, so shorter inputs are returned unchanged.
func shortHash(hash string) string {
	if len(hash) > 7 {
		return hash[:7]
	}
	return hash
}

// subject returns the first line of a commit message, so the table never
// renders a multi-line message.
func subject(message string) string {
	subject, _, _ := strings.Cut(message, "\n")
	return strings.TrimSpace(subject)
}

// relative renders a timestamp as a coarse age. A zero time means the event has
// never happened for this user.
func relative(t time.Time) string {
	if t.IsZero() {
		return "never"
	}

	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}
}

// New creates an empty user detail panel with a configured (but empty) commit
// table.
func New() Model {
	cols := []table.Column{
		{Title: "HASH", Width: hashColWidth},
		{Title: "DATE", Width: dateColWidth},
		{Title: "MESSAGE", Width: 20},
	}

	t := table.New(
		table.WithColumns(cols),
		table.WithRows(nil),
		table.WithFocused(true),
	)

	t.SetStyles(table.Styles{
		Cell: lipgloss.NewStyle().Padding(0, 1),
		// Selected: lipgloss.NewStyle().Foreground(ui.Palette.Accent).Bold(true),
		Header: lipgloss.NewStyle().
			Foreground(ui.Palette.FgMuted).
			Italic(true).
			Padding(0, 1).
			// Border(lipgloss.MarkdownBorder(), false, false, true, false).
			BorderForeground(ui.Palette.SelectedBorder),
	})

	return Model{table: t}
}
