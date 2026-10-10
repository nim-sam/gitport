package serverstatus

/*
 * The server status model is the inline Terminal User Interface (TUI) shown by
 * `gitport start`. It is a pure view: it renders the server's address and
 * alias, a status line and a scrolling viewport that streams the server's logs.
 *
 * All lifecycle work (database bootstrap, migrations, serving the server and
 * reporting its status) is owned by the cli package, which feeds this model
 * through tea messages.
 */

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/nim-sam/gitport/internal/logger"
	"github.com/nim-sam/gitport/internal/mdns"
	"github.com/nim-sam/gitport/internal/model"
	"github.com/nim-sam/gitport/internal/util"
	"github.com/nim-sam/gitport/ui"

	"github.com/charmbracelet/lipgloss/table"
)

// Server lifecycle state, shown in the header.
type status int

const (
	statusStarting status = iota
	statusRunning
	statusStopped
	statusError
)

// Info is the static, display-only information the header renders. It is
// computed by the caller so this model stays decoupled from the server.
type Info struct {
	Name string
	Root string
	Host string
	Port string
}

// Model is the server status TUI's bubbletea model.
type Model struct {
	info Info

	lines []string

	status status
	err    error

	logs viewport.Model
	help help.Model

	width  int
	height int
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "ctrl+q":
			return m, tea.Quit
		case "?":
			m.help.ShowAll = !m.help.ShowAll
		}

	case logger.LogMsg:
		m.appendLog(msg.GetLog())
		return m, nil

	case RunningMsg:
		m.status = statusRunning
		return m, nil

	case StoppedMsg:
		if msg.Err != nil {
			m.status = statusError
			m.err = msg.Err
			return m, nil
		}
		m.status = statusStopped
		return m, tea.Quit
	}

	// The viewport scrolls regardless of the lifecycle state.
	var cmd tea.Cmd
	m.logs, cmd = m.logs.Update(msg)
	_, _ = m.help.Update(msg)
	return m, cmd
}

func (m Model) View() string {

	// Header: display name, alias, address and connect command.

	statusStyle := lipgloss.NewStyle().Bold(true).Foreground(statusColor(m.status))

	vSeparator := lipgloss.NewStyle().Foreground(ui.Palette.FgMuted).Bold(true).Render(" | ")

	header := lipgloss.NewStyle().Width(m.width-4).Padding(0, 1).Render(lipgloss.JoinHorizontal(
		lipgloss.Left,
		lipgloss.NewStyle().Bold(true).Foreground(ui.Palette.Accent).Render("GitPort"),
		vSeparator,
		lipgloss.NewStyle().Foreground(ui.Palette.Fg).Italic(true).Bold(true).Render(m.info.Name),
		lipgloss.NewStyle().Foreground(ui.Palette.FgMuted).Italic(true).Render(" @ "+m.info.Root),
	))

	serverInfo := lipgloss.NewStyle().Width(m.width-2).Padding(0, 1).Render(
		lipgloss.JoinVertical(
			lipgloss.Top,
			m.meta("alias", m.alias()),
			m.meta("address", m.address()),
			m.meta("connect", m.connect()),
			m.meta("status", statusStyle.Render(statusLabel(m.status))),
		),
	)

	// Log viewport.

	logBoxStyle := lipgloss.NewStyle().
		Width(m.width).
		Height(5).
		Padding(0, 1).
		BorderForeground(ui.Palette.Border)

	// composition

	contentContainer := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderRow(true).
		BorderStyle(lipgloss.NewStyle().Foreground(ui.Palette.Border)).
		Width(m.width+2).
		Height(m.height).
		Rows(
			[]string{
				header,
			},
			[]string{
				serverInfo,
			},
			[]string{
				logBoxStyle.Render(m.logs.View()),
			},
		)

	return lipgloss.JoinVertical(lipgloss.Left,
		contentContainer.Render(),
	)
}

// Rendering helpers.

func (m Model) meta(label, value string) string {
	labelStyle := lipgloss.NewStyle().Foreground(ui.Palette.FgMuted).Width(10)
	valueStyle := lipgloss.NewStyle().Foreground(ui.Palette.Fg)
	return labelStyle.Render(label) + valueStyle.Render(value)
}

func (m *Model) appendLog(entry model.Log) {
	line := fmt.Sprintf("%s  %s  %s",
		lipgloss.NewStyle().Foreground(ui.Palette.FgMuted).Render(clock(entry.Date)),
		levelStyle(entry.Level).Render(levelLabel(entry.Level)),
		lipgloss.NewStyle().Foreground(ui.Palette.Fg).Render(entry.Message),
	)

	m.lines = append(m.lines, line)
	m.logs.SetContent(strings.Join(m.lines, "\n"))
	m.logs.GotoBottom()
}

// Address helpers.

func (m Model) alias() string {
	return mdns.Label(m.info.Name) + ".local"
}

func (m Model) address() string {
	host := m.info.Host
	if host == "" || host == "0.0.0.0" {
		host = util.GetLocalIP()
	}
	port := m.info.Port
	if port == "" {
		port = "22"
	}
	return fmt.Sprintf("%s:%s", host, port)
}

func (m Model) connect() string {
	port := m.info.Port
	if port == "" || port == "22" {
		return "ssh " + m.alias()
	}
	return fmt.Sprintf("ssh -p %s %s", port, m.alias())
}

// New creates the server status model. It is a pure view over info; the caller
// owns the lifecycle and feeds it RunningMsg/StoppedMsg as the server changes
// state.
func New(info Info, width, height int) Model {
	return Model{
		info:   info,
		logs:   viewport.New(width-2, 5),
		height: height,
		width:  width,
	}
}

// Styling helpers.

var errorColor = lipgloss.Color("#ff5555")

func levelStyle(level model.LogLevel) lipgloss.Style {
	switch level {
	case model.Error:
		return lipgloss.NewStyle().Foreground(errorColor).Bold(true).Width(7)
	case model.Warning:
		return lipgloss.NewStyle().Foreground(ui.Palette.FgHighlight).Bold(true).Width(7)
	default:
		return lipgloss.NewStyle().Foreground(ui.Palette.FgMuted).Width(7)
	}
}

func levelLabel(level model.LogLevel) string {
	switch level {
	case model.Error:
		return "ERROR"
	case model.Warning:
		return "WARN"
	default:
		return "INFO"
	}
}

func statusLabel(s status) string {
	switch s {
	case statusRunning:
		return "[ Running ]"
	case statusStopped:
		return "[ Stopped ]"
	case statusError:
		return "[ Error ]"
	default:
		return "[ Starting ]"
	}
}

func statusColor(s status) lipgloss.Color {
	switch s {
	case statusRunning:
		return ui.Palette.Accent
	case statusError:
		return errorColor
	case statusStopped:
		return ui.Palette.FgMuted
	default:
		return ui.Palette.FgHighlight
	}
}

// clock trims a stored "2006/01/02 15:04:05" timestamp down to its time part.
func clock(date string) string {
	if len(date) >= 19 {
		return date[11:]
	}
	return date
}
