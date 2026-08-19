package components

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nim-sam/gitport/pkg/selector"
	"github.com/nim-sam/gitport/ui"
)

type UserFormModel struct {
	Inputs        []textinput.Model
	selectedInput int
	Selector      selector.Model
	Width         int
	Height        int
	buttonMsg     string
	title         string
}

func (m UserFormModel) Init() tea.Cmd {
	return nil
}

func (m UserFormModel) Update(msg tea.Msg) (UserFormModel, tea.Cmd) {
	for i, _ := range m.Inputs {
		m.Inputs[i].Width = m.Width - 18
	}
	return m, nil
}

func (m UserFormModel) View() string {

	var formFieldsView []string

	// Styling form title
	titleStyle := lipgloss.NewStyle().
		Width(m.Width-2).
		Bold(true).
		Foreground(ui.Palette.FgHighlight).
		AlignHorizontal(lipgloss.Center).
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(ui.Palette.Border)

	// Styling text input models
	for _, model := range m.Inputs {

		model.TextStyle = model.TextStyle.Foreground(ui.Palette.Fg)
		model.PlaceholderStyle = model.PlaceholderStyle.Foreground(ui.Palette.FgMuted).Italic(true)
		model.Cursor.Style = model.Cursor.Style.Foreground(ui.Palette.FgHighlight)
		model.PromptStyle = model.PromptStyle.Foreground(ui.Palette.FgHighlight)

		formFieldsView = append(formFieldsView, model.View())
	}

	// Styling submit button
	focusedButton := lipgloss.NewStyle().Foreground(ui.Palette.FgHighlight).Bold(true)
	blurredButton := lipgloss.NewStyle().Foreground(ui.Palette.FgMuted)

	var button lipgloss.Style
	if m.selectedInput == len(m.Inputs)+1 {
		button = focusedButton
	} else {
		button = blurredButton
	}

	// Styling Selector carrousel
	focusedSelector := lipgloss.NewStyle().Foreground(ui.Palette.Fg)
	blurredSelector := lipgloss.NewStyle().Foreground(ui.Palette.FgMuted).Italic(true)

	var Selector lipgloss.Style
	if m.selectedInput == len(m.Inputs) {
		Selector = focusedSelector
	} else {
		Selector = blurredSelector
	}

	m.Selector.Styles.Carrousel = Selector

	// Styling Selector and text input prompt
	formPromptStyle := lipgloss.NewStyle().Bold(true).Foreground(ui.Palette.FgHighlight)
	m.Selector.Styles.Prompt = formPromptStyle

	for i, _ := range m.Inputs {
		m.Inputs[i].PromptStyle = formPromptStyle
	}

	// Composing the form
	titleBar := lipgloss.NewStyle().Width(m.Width).AlignHorizontal(lipgloss.Center)
	composition := lipgloss.NewStyle().Width(m.Width-2).Height(m.Height-5).Padding(0, 1).AlignVertical(lipgloss.Center).AlignHorizontal(lipgloss.Center)
	inputsContainer := lipgloss.NewStyle().Width(m.Width-12).Height(3).Padding(0, 1)
	buttonContainer := lipgloss.NewStyle().Width(m.Width).Height(3).Align(lipgloss.Center, lipgloss.Center)

	return lipgloss.JoinVertical(lipgloss.Top,
		titleBar.Render(titleStyle.Render(m.title)),
		composition.Render(
			inputsContainer.Render(
				lipgloss.JoinVertical(lipgloss.Top,
					strings.Join(formFieldsView, "\n"),
					m.Selector.View(),
				),
			),
		),
		buttonContainer.Render(button.Render(m.buttonMsg)),
	)
}

func NewUserCreationFormModel() UserFormModel {

	s := selector.New("ReadOnly", "ReadWrite", "Admin")

	s.Prompt = "◇"

	u := UserFormModel{
		Inputs:    make([]textinput.Model, 2),
		buttonMsg: "[ SUBMIT ]",
		title:     "Create New User",
		Selector:  s,
	}

	for i := range u.Inputs {

		t := textinput.New()

		switch i {
		case 0:
			t.Prompt = "◇ "
			t.CharLimit = 32
			t.Placeholder = "Username"
			// t.Width = u.width - 4
			t.Focus()
		case 1:
			t.Prompt = "◇ "
			t.CharLimit = 512
			t.Placeholder = "SSH Key"
		}

		u.Inputs[i] = t

	}

	return u

}

func NewUserModificationFormModel() UserFormModel {

	s := selector.New("ReadOnly", "ReadWrite", "Admin")

	s.Prompt = "◇"

	u := UserFormModel{
		Inputs:    make([]textinput.Model, 2),
		buttonMsg: "[ CONFIRM ]",
		title:     "Edit User Information",
		Selector:  s,
	}

	for i := range u.Inputs {

		t := textinput.New()

		switch i {
		case 0:
			t.Prompt = "◇ "
			t.CharLimit = 32
			t.Placeholder = "Username"
			// t.Width = u.width - 4
			t.Focus()
		case 1:
			t.Prompt = "◇ "
			t.CharLimit = 512
			t.Placeholder = "SHH Key"
		}

		u.Inputs[i] = t

	}

	return u

}

// Handles key Inputs to navigate and interact with UserFormModel's fields
func (m *UserFormModel) UpdateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.Inputs))

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up":
			if m.selectedInput > 0 {
				m.selectedInput--
			}
		case "down", "enter":
			if m.selectedInput < len(m.Inputs)+1 {
				m.selectedInput++
			}
		}
	}

	for i := range len(m.Inputs) {
		if i == m.selectedInput {
			m.Inputs[i].Focus()
		} else {
			m.Inputs[i].Blur()
		}
	}

	if m.selectedInput == 2 {
		m.Selector.Focused = true
	} else {
		m.Selector.Focused = false
	}

	for i := range m.Inputs {
		m.Inputs[i], cmds[i] = m.Inputs[i].Update(msg)
	}

	m.Selector.UpdateSelection(msg)

	return tea.Batch(cmds...)
}

func (m *UserFormModel) ClearInputs() {

	for i := range m.Inputs {
		m.Inputs[i].SetValue("")
	}

}
