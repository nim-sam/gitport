package dialog

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type selectedButton int

const (
	confirm selectedButton = iota
	cancel
)

type SelectionMsg struct {
	Confirm bool
}

// Dialog Model struct
type Model struct {
	Title      string // Title of message window
	Prompt     string // Content of the message
	ConfirmMsg string // Msg of button confirming action
	CancelMsg  string // Msg of button canceling action
	Style      Style  // Dialog styles

	// Dimensions
	Width  int
	Height int

	selectedButton selectedButton
}

// Dialog Model styles
type Style struct {
	Title              lipgloss.Style
	Prompt             lipgloss.Style
	ConfirmMsgNormal   lipgloss.Style
	CancelMsgNormal    lipgloss.Style
	ConfirmMsgSelected lipgloss.Style
	CancelMsgSelected  lipgloss.Style
}

// Implementing bubble tea model interface
func (m Model) Init() tea.Cmd { return nil }
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {

	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		// Switch between button selection using the tab or left/right arrow buttons
		case "tab":
			switch m.selectedButton {
			case confirm:
				m.selectedButton = cancel
			case cancel:
				m.selectedButton = confirm
			}
		case "left", "h":
			m.selectedButton = cancel
		case "right", "l":
			m.selectedButton = confirm
		// Send button selection through tea command
		case "enter":
			cmd = func() tea.Msg { return SelectionMsg{Confirm: m.selectedButton == confirm} }
		}
	}

	return m, cmd
}

func (m Model) View() string {

	// Containers
	globalContainer := lipgloss.NewStyle().
		Height(m.Height).
		Width(m.Width).
		Padding(0, 1)
	buttonContainer := lipgloss.NewStyle().
		Height(1).
		AlignHorizontal(lipgloss.Center).
		Width(m.Width)

	// Style rendering
	title := m.Style.Title.Width(m.Width - 2).
		AlignHorizontal(lipgloss.Center).
		// Border(lipgloss.HiddenBorder(), false, false, true, false).
		Height(1).
		Render(m.Title)
	prompt := m.Style.Prompt.Width(m.Width).
		AlignHorizontal(lipgloss.Center).
		AlignVertical(lipgloss.Center).
		Height(m.Height - 5).
		Render(m.Prompt)

	// Button selection
	var confirmButton string
	var cancelButton string

	switch m.selectedButton {
	case confirm:
		confirmButton = m.Style.ConfirmMsgSelected.Render(m.ConfirmMsg)
		cancelButton = m.Style.CancelMsgNormal.Render(m.CancelMsg)
	case cancel:
		confirmButton = m.Style.ConfirmMsgNormal.Render(m.ConfirmMsg)
		cancelButton = m.Style.ConfirmMsgSelected.Render(m.CancelMsg)
	}

	buttonsRendered := buttonContainer.Render(
		lipgloss.JoinHorizontal(
			lipgloss.Left,
			cancelButton,
			"   ",
			confirmButton,
		),
	)

	dialogRendered := globalContainer.Render(
		lipgloss.JoinVertical(
			lipgloss.Top,
			title,
			prompt,
			buttonsRendered,
		),
	)

	return dialogRendered
}

// Selection getter
func (m Model) Selection() selectedButton {
	return m.selectedButton
}

// Resets Selection
func (m *Model) ResetSelection() {
	m.selectedButton = cancel
}

// Creating new dialog
func New() Model {
	return Model{}
}
