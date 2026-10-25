package commitgraph

import (
	"fmt"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
	"github.com/NimbleMarkets/ntcharts/v2/canvas/runes"
	"github.com/NimbleMarkets/ntcharts/v2/linechart/timeserieslinechart"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/nim-sam/gitport/internal/model"
	"github.com/nim-sam/gitport/ui"
)

const (
	maxWeeks = 26 // how far back the graph goes
	minWeeks = 8  // fewest weeks shown on a very narrow panel

	// yAxisMargin is the columns the library uses for the Y axis and labels.
	// It's a guess; raise it if the right end of the line is clipped.
	yAxisMargin = 8

	// Placeholder titles.
	chartTitle = "Commit Activity"
	yAxisTitle = "Commits"
	xAxisTitle = "Week"

	// yGutter is the columns reserved left of the chart for the Y axis title.
	yGutter = 2

	// chartHeight is the preferred height of the graph itself. It is capped to
	// the room the table leaves for it.
	chartHeight    = 10
	minChartHeight = 4

	// Table geometry. The hidden border adds 2 columns (left/right) and 4 rows
	// (top, bottom, and one separator between each of the 3 rows).
	tableBorderCols = 2
	tableBorderRows = 4
	titleRows       = 1
	statsRows       = 2
	xTitleRows      = 1

	minStatWidth = 12

	defaultOuterWidth  = 70
	defaultOuterHeight = 20

	graphPadX = 2
)

var (
	lineStyle  = lipgloss.NewStyle().Foreground(ui.Palette.Accent)
	axisStyle  = lipgloss.NewStyle().Foreground(ui.Palette.FgMuted)
	labelStyle = lipgloss.NewStyle().Foreground(ui.Palette.Fg)

	titleStyle     = lipgloss.NewStyle().Foreground(ui.Palette.FgHighlight).Bold(true)
	axisTitleStyle = lipgloss.NewStyle().Foreground(ui.Palette.FgMuted).Italic(true)

	statLabelStyle = lipgloss.NewStyle().Foreground(ui.Palette.FgMuted)
	statValueStyle = lipgloss.NewStyle().Foreground(ui.Palette.Fg).Bold(true)
)

// stat is one labelled value shown under the graph.
type stat struct{ label, value string }

type Model struct {
	commits []model.Commit
	chart   timeserieslinechart.Model
	stats   []stat

	// Outer size of the whole widget, hidden table border included.
	outerWidth  int
	outerHeight int

	// Width of the content area inside the table's hidden border.
	innerWidth int

	// Height of the middle table row (graph + its X axis title).
	midHeight int

	// Size of the ntcharts chart itself.
	width  int
	height int
}

func New(commits []model.Commit) Model {
	m := Model{commits: commits}
	m.SetSize(defaultOuterWidth, defaultOuterHeight)
	return m
}

// SetSize sets the OUTER size of the widget and derives the chart size from
// it. The ntcharts model is fixed-size, so it is rebuilt.
func (m *Model) SetSize(width, height int) {
	m.outerWidth, m.outerHeight = width, height

	m.innerWidth = max(width-tableBorderCols, minStatWidth*3)

	// Rows left for the middle table row once the title, stats and borders
	// have taken theirs.
	m.midHeight = max(height-tableBorderRows-titleRows-statsRows, minChartHeight+xTitleRows)

	m.width = m.innerWidth - yGutter - 2*graphPadX
	m.height = max(min(chartHeight, m.midHeight-xTitleRows), minChartHeight)

	m.rebuild()
}

// SetCommits replaces the data, e.g. after the repo gets new commits.
func (m *Model) SetCommits(commits []model.Commit) {
	m.commits = commits
	m.rebuild()
}

// weekStart returns local midnight of the Monday of t's week.
func weekStart(t time.Time) time.Time {
	y, mo, d := t.Local().Date()
	day := time.Date(y, mo, d, 0, 0, 0, 0, time.Local)
	offset := (int(day.Weekday()) + 6) % 7 // Monday = 0
	return day.AddDate(0, 0, -offset)
}

// weeksForWidth picks one week per plot column, within [minWeeks, maxWeeks].
func (m Model) weeksForWidth() int {
	return min(maxWeeks, max(minWeeks, m.width-yAxisMargin))
}

// weeklyCounts returns one point for EVERY week in the window (zero for empty
// weeks), so the X range is always the same fixed span regardless of repo age.
func weeklyCounts(commits []model.Commit, weeks int) (points []timeserieslinechart.TimePoint, maxCount float64, from, to time.Time) {
	thisWeek := weekStart(time.Now())
	from = thisWeek.AddDate(0, 0, -(weeks-1)*7)
	to = thisWeek

	counts := make(map[time.Time]int)
	for _, c := range commits {
		w := weekStart(c.Date)
		if w.Before(from) || w.After(thisWeek) {
			continue
		}
		counts[w]++
	}

	for w := from; !w.After(thisWeek); w = w.AddDate(0, 0, 7) {
		v := float64(counts[w])
		maxCount = max(maxCount, v)
		points = append(points, timeserieslinechart.TimePoint{Time: w, Value: v})
	}
	return points, maxCount, from, to
}

// ago renders a time as a short relative duration.
func ago(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	case d < 60*24*time.Hour:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	default:
		return fmt.Sprintf("%dmo ago", int(d.Hours()/(24*30)))
	}
}

// computeStats derives the three headline numbers shown under the graph:
// activity this week vs last, freshness of the repo, and the top contributor.
func computeStats(commits []model.Commit) []stat {
	if len(commits) == 0 {
		return []stat{
			{"This week", "0"},
			{"Last commit", "never"},
			{"Top author", "none"},
		}
	}

	thisWeek := weekStart(time.Now())
	lastWeek := thisWeek.AddDate(0, 0, -7)

	var thisN, lastN int
	var last time.Time
	authors := make(map[string]int)

	for _, c := range commits {
		authors[c.Author]++
		if c.Date.After(last) {
			last = c.Date
		}

		w := weekStart(c.Date)
		if w.Equal(thisWeek) {
			thisN++
		} else if w.Equal(lastWeek) {
			lastN++
		}
	}

	topAuthor, topN := "", 0
	for a, n := range authors {
		if n > topN || (n == topN && a < topAuthor) { // alphabetical tie-break
			topAuthor, topN = a, n
		}
	}

	return []stat{
		{"This week", fmt.Sprintf("%d (%+d vs last)", thisN, thisN-lastN)},
		{"Last commit", ago(last)},
		{"Top author", fmt.Sprintf("%s (%d)", topAuthor, topN)},
	}
}

func (m *Model) rebuild() {
	points, maxCount, from, to := weeklyCounts(m.commits, m.weeksForWidth())
	maxCount = max(maxCount, 1)

	chart := timeserieslinechart.New(m.width, m.height,
		timeserieslinechart.WithTimeRange(from, to), // fixed X window, in weeks
		timeserieslinechart.WithYRange(0, maxCount),
		timeserieslinechart.WithAxesStyles(axisStyle, labelStyle),
		timeserieslinechart.WithStyle(lineStyle),
		timeserieslinechart.WithLineStyle(runes.ArcLineStyle), // rounded corners
	)

	for _, p := range points {
		chart.Push(p)
	}

	chart.DrawXYAxisAndLabel()
	chart.Draw()
	m.chart = chart
	m.stats = computeStats(m.commits)
}

// yTitleView renders the Y axis title vertically (one letter per row),
// centered next to the chart. It is cut short if the chart is too short.
func (m Model) yTitleView() string {
	letters := []rune(yAxisTitle)
	if len(letters) > m.height {
		letters = letters[:m.height]
	}
	parts := make([]string, len(letters))
	for i, r := range letters {
		parts[i] = string(r)
	}

	return axisTitleStyle.
		Width(yGutter).
		Height(m.height).
		AlignVertical(lipgloss.Center).
		Render(strings.Join(parts, "\n"))
}

// graphView is the middle table row: Y axis title, chart, and X axis title.
func (m Model) graphView() string {
	body := lipgloss.JoinHorizontal(lipgloss.Top, m.yTitleView(), m.chart.View())

	// X axis title, centered under the chart (not under the Y gutter).
	xTitle := strings.Repeat(" ", yGutter) + axisTitleStyle.
		Width(m.width).
		AlignHorizontal(lipgloss.Center).
		Render(xAxisTitle)

	return lipgloss.JoinVertical(lipgloss.Left, body, xTitle)
}

// statsView lays the stats out in a single row of equal columns, each stat
// centered in its column.
func (m Model) statsView() string {
	if len(m.stats) == 0 {
		return ""
	}

	cell := lipgloss.NewStyle().
		Width(max(m.innerWidth/len(m.stats), minStatWidth)).
		AlignHorizontal(lipgloss.Center)

	cells := make([]string, 0, len(m.stats))
	for _, s := range m.stats {
		cells = append(cells, cell.Render(
			statLabelStyle.Render(s.label)+"\n"+statValueStyle.Render(s.value)))
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, cells...)
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) { return m, nil }

func (m Model) View() string {
	t := table.New().
		// Hidden border: blank characters that only provide margins and the
		// gaps between rows.
		Border(lipgloss.HiddenBorder()).
		BorderRow(true).
		Width(m.outerWidth).
		StyleFunc(func(row, col int) lipgloss.Style {
			s := lipgloss.NewStyle().AlignHorizontal(lipgloss.Center)
			if row == 1 {
				s = s.
					Height(m.midHeight).
					AlignVertical(lipgloss.Center).
					Padding(0, graphPadX) // left and right only
			}
			return s
		}).
		Rows(
			[]string{titleStyle.Render(chartTitle)},
			[]string{m.graphView()},
			[]string{m.statsView()},
		)

	return t.Render()
}
