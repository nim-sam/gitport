package serveroverview

import (
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	lipglosstable "github.com/charmbracelet/lipgloss/table"
	"github.com/nim-sam/gitport/internal/config"
	"github.com/nim-sam/gitport/ui"
)

const (
	// labelWidth is the width of the static setting column shared by the
	// view and edit modes.
	labelWidth = 20

	// valueGutter is the horizontal space reserved around a text input inside
	// the settings table (label column and container borders), so the input's
	// viewport never exceeds its dedicated area.
	valueGutter = 24
)

// settings describes the config fields shown by the overview, in display and
// focus order.
var settings = []struct {
	label string
	limit int
}{
	{"Display Name", 128},
	{"Binding Address", 128},
	{"Port", 5},
}

var (
	settingStyle = lipgloss.NewStyle().Height(1).Padding(0, 1).Foreground(ui.Palette.Fg)
	labelStyle   = lipgloss.NewStyle().Foreground(ui.Palette.Fg).Bold(true).Width(labelWidth)
)

type Model struct {
	width         int
	height        int
	root          string
	config        config.Config
	settingsTable table.Model
	inputs        []textinput.Model

	// editable is injected by the dashboard for Admins and Owners. When false
	// ctrl+e is ignored and the footer reports the view as read-only.
	editable bool

	// editing is true while the config values are rendered as text inputs.
	editing       bool
	selectedInput int
	errMsg        string
	statusMsg     string
}

func New() Model {
	root, err := os.Getwd()
	if err != nil {
		root = "."
	}

	cfg, err := config.Load(root)
	if err != nil {
		cfg = config.Config{}
	}

	m := Model{
		root:   root,
		config: cfg,
		inputs: newInputs(),
	}
	return m
}

func newInputs() []textinput.Model {
	inputs := make([]textinput.Model, len(settings))
	for i, setting := range settings {
		input := textinput.New()
		input.Prompt = ""
		input.CharLimit = setting.limit
		inputs[i] = input
	}
	return inputs
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Width(w int) {
	m.width = w
}

func (m *Model) Height(h int) {
	m.height = h

}

// SetEditable injects the current session's permission to change the config.
// Losing the permission while editing cancels the pending changes.
func (m *Model) SetEditable(editable bool) {
	m.editable = editable
	if !editable && m.editing {
		m.cancelEdit()
	}
}

// Editing reports whether the config fields are currently rendered as inputs.
// The dashboard uses it to suppress tab switching while changes are pending.
func (m *Model) Editing() bool {
	return m.editing
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {

	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "ctrl+e":
			if !m.editable {
				return m, nil
			}
			if m.editing {
				m.save()
			} else {
				m.beginEdit()
			}
			return m, nil
		}

		if m.editing {
			switch key.String() {
			case "esc":
				m.cancelEdit()
			case "enter", "down":
				m.focusInput(m.selectedInput + 1)
			case "up":
				m.focusInput(m.selectedInput - 1)
			default:
				var cmd tea.Cmd
				m.inputs[m.selectedInput], cmd = m.inputs[m.selectedInput].Update(msg)
				return m, cmd
			}
			return m, nil
		}
	}

	if m.editing {
		// Non-key messages (e.g. cursor blink) are broadcast to every input;
		// each cursor ignores blinks addressed to another input.
		var cmds tea.BatchMsg
		for i := range m.inputs {
			var cmd tea.Cmd
			m.inputs[i], cmd = m.inputs[i].Update(msg)
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)
	}

	return m, nil
}

func (m Model) View() string {

	contentContainers := lipgloss.NewStyle().Width(m.width).Height(m.height - 6).AlignHorizontal(lipgloss.Center).AlignVertical(lipgloss.Center)

	content := m.displayView()
	if m.editing {
		content = m.inputsView()
	}

	contentTable := lipglosstable.New().Border(
		lipgloss.RoundedBorder(),
	).BorderStyle(
		lipgloss.NewStyle().Foreground(ui.Palette.Border),
	).BorderRow(
		true,
	).Width(
		m.width+2,
	).Height(
		m.height,
	).Rows(
		[]string{
			lipgloss.NewStyle().
				Bold(true).
				AlignHorizontal(lipgloss.Center).
				Width(m.width).
				Foreground(ui.Palette.FgHighlight).
				Render("Settings"),
		},
		[]string{
			contentContainers.Render(content),
		},
		[]string{
			m.footerView(),
		},
	)

	return contentTable.Render()
}

// displayView renders the three config fields as minimal table
func (m Model) displayView() string {
	settingValues := []string{m.config.DisplayName, m.config.Host, m.config.Port}
	rows := make([]string, 0, len(m.inputs))
	for i := range m.inputs {
		rows = append(rows, lipgloss.JoinHorizontal(
			lipgloss.Left,
			labelStyle.Render(settings[i].label),
			settingStyle.Render(settingValues[i]),
		))
	}
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

// inputsView renders the three config fields as labelled text inputs.
func (m Model) inputsView() string {
	rows := make([]string, 0, len(m.inputs))
	for i := range m.inputs {
		input := styleForView(m.inputs[i])
		rows = append(rows, lipgloss.JoinHorizontal(
			lipgloss.Left,
			labelStyle.Render(settings[i].label),
			settingStyle.Render(input.View()),
		))
	}
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func (m Model) footerView() string {
	hintStyle := lipgloss.NewStyle().Foreground(ui.Palette.FgMuted).Width(m.width).Height(3).AlignHorizontal(lipgloss.Center).AlignVertical(lipgloss.Center)

	switch {
	case m.errMsg != "":
		return lipgloss.NewStyle().Foreground(ui.Palette.Error).Render(m.errMsg)
	case m.statusMsg != "":
		return hintStyle.Render(m.statusMsg)
	case m.editing:
		return hintStyle.Render("Press ctrl+e to save settings.")
	case !m.editable:
		return hintStyle.Render("Read-Only. Contact your administrator for any modification requests.")
	default:
		return hintStyle.Render("Server configs in normal mode.")
	}
}

// styleForView applies the palette to a copy of a text input so its value and
// cursor render in the theme's colors.
func styleForView(model textinput.Model) textinput.Model {
	model.TextStyle = model.TextStyle.Foreground(ui.Palette.Fg)
	model.PlaceholderStyle = model.PlaceholderStyle.Foreground(ui.Palette.FgMuted).Italic(true)
	model.Cursor.Style = model.Cursor.Style.Foreground(ui.Palette.FgHighlight)
	// The char under a blurred cursor renders via Cursor.TextStyle, which
	// bubbles leaves at its zero value; without this it shows the terminal's
	// default color instead of the theme's text color.
	model.Cursor.TextStyle = model.TextStyle
	model.PromptStyle = model.PromptStyle.Foreground(ui.Palette.FgHighlight)
	return model
}

func (m *Model) beginEdit() {
	m.editing = true
	m.errMsg = ""
	m.statusMsg = ""
	m.selectedInput = 0

	// m.syncWidths()
	values := []string{m.config.DisplayName, m.config.Host, m.config.Port}
	for i := range m.inputs {
		m.inputs[i].SetValue(values[i])
		// SetValue leaves the cursor at the end; show overlong values from the
		// head so they fit the field without scrolling on first render.
		m.inputs[i].CursorStart()
	}
	m.focusInput(0)
}

func (m *Model) cancelEdit() {
	m.editing = false
	m.errMsg = ""
	m.statusMsg = ""
	for i := range m.inputs {
		m.inputs[i].Blur()
	}
}

// save validates the inputs and persists them. On failure the form stays open
// with the error in the footer; on success the view mode is restored.
func (m *Model) save() {
	cfg := config.Config{
		DisplayName: strings.TrimSpace(m.inputs[0].Value()),
		Host:        strings.TrimSpace(m.inputs[1].Value()),
		Port:        strings.TrimSpace(m.inputs[2].Value()),
	}

	if errMsg := validate(cfg); errMsg != "" {
		m.errMsg = errMsg
		return
	}

	if err := config.Save(m.root, cfg); err != nil {
		m.errMsg = err.Error()
		return
	}

	// Reload so the defaults applied by Save are reflected in the table.
	reloaded, err := config.Load(m.root)
	if err != nil {
		m.errMsg = err.Error()
		return
	}

	m.config = reloaded
	m.editing = false
	m.errMsg = ""
	m.statusMsg = "Saved. Restart the server to apply."
	for i := range m.inputs {
		m.inputs[i].Blur()
	}
}

func validate(cfg config.Config) string {
	if cfg.Port == "" {
		return ""
	}
	port, err := strconv.Atoi(cfg.Port)
	if err != nil || port < 1 || port > 65535 {
		return "Port must be a number between 1 and 65535"
	}
	return ""
}

func (m *Model) focusInput(index int) {
	if len(m.inputs) == 0 {
		return
	}

	count := len(m.inputs)
	m.selectedInput = ((index % count) + count) % count
	for i := range m.inputs {
		if i == m.selectedInput {
			m.inputs[i].Focus()
		} else {
			m.inputs[i].Blur()
		}
	}
}
