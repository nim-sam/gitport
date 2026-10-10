package initprompter

/*
 * The init prompter is the inline Terminal User Interface (TUI) shown by
 * `gitport init`. It walks the user through the server's starter configuration
 * one field at a time - display name, bind address, port and the single admin
 * account - then confirms before the handler persists it.
 *
 * It renders inline (no alternate screen) so the wizard reads like a sequence
 * of shell prompts, matching the command line tool it belongs to.
 */

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/nim-sam/gitport/internal/config"
	"github.com/nim-sam/gitport/internal/model"
	"github.com/nim-sam/gitport/pkg/selector"
	"github.com/nim-sam/gitport/ui"

	"github.com/charmbracelet/lipgloss/table"
)

// step enumerates the wizard's prompts in order.
type step int

const (
	stepDisplayName step = iota
	stepHost
	stepPort
	stepAdminName
	stepAdminEmail
	stepReview
)

const stepCount = int(stepReview) + 1

// Model is the init wizard's bubbletea model.
type Model struct {
	inputs []textinput.Model
	access selector.Model
	step   step
	width  int
	height int
	result *InitConfigMsg
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "ctrl+q":
			return m, tea.Quit
		case "esc":
			if m.step == stepDisplayName {
				return m, tea.Quit
			}
			m.step--
		case "enter":
			if m.step == stepReview {
				cfg, admin := m.assemble()
				m.result = &InitConfigMsg{Config: cfg, Admin: admin}
				return m, tea.Quit
			}
			m.step++
		}
	}

	// Route the message to whichever field is active, keeping every other
	// field blurred so its cursor never shows.
	switch m.step {
	case stepReview:
		m.blurInputs()
		m.access.Focused = false
	default:
		m.access.Focused = false
		m.focusInput(int(m.step))
		var cmd tea.Cmd
		m.inputs[int(m.step)], cmd = m.inputs[int(m.step)].Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m Model) View() string {

	// title := lipgloss.NewStyle().Bold(true).Foreground(ui.Palette.Accent).Render("GitPort Setup")
	// progress := lipgloss.NewStyle().Foreground(ui.Palette.FgMuted).Render(
	// 	fmt.Sprintf("Step %d/%d", int(m.step)+1, stepCount))

	title := lipgloss.NewStyle().AlignHorizontal(lipgloss.Left).Width(m.width/3-1).Bold(true).Foreground(ui.Palette.Accent).Padding(0, 1).Render("Gitport")
	currentStep := lipgloss.NewStyle().AlignHorizontal(lipgloss.Center).Width(m.width / 3).Foreground(ui.Palette.Fg).Render(stepAsString(m.step))
	progress := lipgloss.NewStyle().AlignHorizontal(lipgloss.Right).Width(m.width/3 - 1).Foreground(ui.Palette.FgMuted).Render(
		fmt.Sprintf("%d/%d", int(m.step)+1, stepCount),
	)

	header := lipgloss.NewStyle().Width(m.width - 2).Render(
		lipgloss.JoinHorizontal(
			lipgloss.Left,
			title,
			currentStep,
			progress,
		),
	)

	var body string
	var hint string
	switch m.step {
	case stepDisplayName:
		body = m.inputs[0].View()
		hint = "This is the name of the server instace. You can name it however you want."
	case stepHost:
		body = m.inputs[1].View()
		hint = "Bind to 0.0.0.0 to provide LAN access to the server."
	case stepPort:
		body = m.inputs[2].View()
		hint = "Port 22 requires elevated permissions when running the server."
	case stepAdminName:
		body = m.inputs[3].View()
		hint = "User name of the root admin. Ideally matches with your Git config.name value."
	case stepAdminEmail:
		body = m.inputs[4].View()
		hint = "Email address of the root admin. Ideally matches with your Git config.email value."
	case stepReview:
		body = m.review()
		hint = "Upon review, press Enter to save the GitPort configs."
	}

	hintStyle := lipgloss.NewStyle().Foreground(ui.Palette.FgMuted).Height(2).Padding(0, 1).Width(m.width - 2)
	hint = hintStyle.Render(hint)

	bodyStyle := lipgloss.NewStyle().Height(4).Padding(0, 1).Width(m.width - 2)
	body = bodyStyle.Render(body)

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
				body,
			},
			[]string{
				hint,
			},
		)

	return contentContainer.Render()
}

// Result returns the gathered configuration, or nil if the wizard was aborted.
func (m Model) Result() *InitConfigMsg {
	return m.result
}

// review summarises the assembled configuration before it is written.
func (m Model) review() string {
	cfg, admin := m.assemble()

	labelStyle := lipgloss.NewStyle().Foreground(ui.Palette.FgMuted).Width(16)
	valueStyle := lipgloss.NewStyle().Foreground(ui.Palette.Fg)

	row := func(label, value string) string {
		return labelStyle.Render(label) + valueStyle.Render(value)
	}

	name := cfg.DisplayName
	if name == "" {
		name = "(repository name)"
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		row("Display name", name),
		row("Bind address", fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)),
		row("Admin", admin.Name),
		row("Email", admin.Email),
	)
}

// assemble builds the configuration and admin user from the current inputs,
// applying the same defaults config.Save would.
func (m Model) assemble() (config.Config, model.User) {
	host := m.inputs[1].Value()
	if host == "" {
		host = "0.0.0.0"
	}
	port := m.inputs[2].Value()
	if port == "" {
		port = "22"
	}

	cfg := config.Config{
		DisplayName: m.inputs[0].Value(),
		Host:        host,
		Port:        port,
	}

	admin := model.User{
		Name:        m.inputs[3].Value(),
		Email:       m.inputs[4].Value(),
		AccessLevel: model.Admin,
	}

	return cfg, admin
}

func (m *Model) focusInput(idx int) {
	for i := range m.inputs {
		if i == idx {
			m.inputs[i].Focus()
		} else {
			m.inputs[i].Blur()
		}
	}
}

func (m *Model) blurInputs() {
	for i := range m.inputs {
		m.inputs[i].Blur()
	}
}

// New creates the init wizard. repoName pre-fills the display name so the
// common case only needs a confirmation.
func New(width int, height int) Model {

	placeholders := []struct {
		placeholder string
		charLimit   int
	}{
		{"Server Display Name", 64},
		{"Binding Address", 64},
		{"Port", 5},
		{"Username", 32},
		{"Email Address", 128},
	}

	inputs := make([]textinput.Model, len(placeholders))
	for i, p := range placeholders {
		t := textinput.New()
		t.Prompt = "○ "
		t.PromptStyle = lipgloss.NewStyle().Foreground(ui.Palette.Fg)
		t.CharLimit = p.charLimit
		t.Placeholder = p.placeholder
		t.TextStyle = lipgloss.NewStyle().Foreground(ui.Palette.Fg)
		t.Cursor.Style = lipgloss.NewStyle().Foreground(ui.Palette.Fg)
		t.Width = width - 4
		inputs[i] = t
	}

	m := Model{
		inputs: inputs,
		width:  width,
		height: height,
	}
	m.focusInput(int(stepDisplayName))

	return m
}

func stepAsString(step step) string {
	switch step {
	case stepDisplayName:
		return "Display Name"
	case stepHost:
		return "Host Setup"
	case stepPort:
		return "Port Setup"
	case stepAdminName:
		return "Admin username"
	case stepAdminEmail:
		return "Admin email"
	case stepReview:
		return "Review"
	}
	return ""
}
