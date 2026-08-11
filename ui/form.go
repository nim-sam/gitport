package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nim-sam/gitport/pkg/selector"
)

type UserFormModel struct {
	inputs        []textinput.Model
	selectedInput int
	selector      selector.Model
	width         int
	heigt         int
	buttonMsg     string
	title         string
}

func (m UserFormModel) Init() tea.Cmd {
	return nil
}

func (m UserFormModel) Update(msg tea.Msg) (UserFormModel, tea.Cmd) {
	for i, _ := range m.inputs {
		m.inputs[i].Width = m.width - 18
	}
	return m, nil
}

func (m UserFormModel) View() string {

	var formFieldsView []string

	// Styling form title
	titleStyle := lipgloss.NewStyle().
		Width(m.width-2).
		Bold(true).
		Foreground(palette.fgHighlight).
		AlignHorizontal(lipgloss.Center).
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(palette.border)

	// Styling text input models
	for _, model := range m.inputs {

		model.TextStyle = model.TextStyle.Foreground(palette.fg)
		model.PlaceholderStyle = model.PlaceholderStyle.Foreground(palette.fgMuted).Italic(true)
		model.Cursor.Style = model.Cursor.Style.Foreground(palette.fgHighlight)
		model.PromptStyle = model.PromptStyle.Foreground(palette.fgHighlight)

		formFieldsView = append(formFieldsView, model.View())
	}

	// Styling submit button
	focusedButton := lipgloss.NewStyle().Foreground(palette.fgHighlight).Bold(true)
	blurredButton := lipgloss.NewStyle().Foreground(palette.fgMuted)

	var button lipgloss.Style
	if m.selectedInput == len(m.inputs)+1 {
		button = focusedButton
	} else {
		button = blurredButton
	}

	// Styling selector carrousel
	focusedSelector := lipgloss.NewStyle().Foreground(palette.fg)
	blurredSelector := lipgloss.NewStyle().Foreground(palette.fgMuted).Italic(true)

	var selector lipgloss.Style
	if m.selectedInput == len(m.inputs) {
		selector = focusedSelector
	} else {
		selector = blurredSelector
	}

	m.selector.Styles.Carrousel = selector

	// Styling selector and text input prompt
	formPromptStyle := lipgloss.NewStyle().Bold(true).Foreground(palette.fgHighlight)
	m.selector.Styles.Prompt = formPromptStyle

	for i, _ := range m.inputs {
		m.inputs[i].PromptStyle = formPromptStyle
	}

	// Composing the form
	titleBar := lipgloss.NewStyle().Width(m.width).AlignHorizontal(lipgloss.Center)
	composition := lipgloss.NewStyle().Width(m.width-2).Height(m.heigt-5).Padding(0, 1).AlignVertical(lipgloss.Center).AlignHorizontal(lipgloss.Center)
	inputsContainer := lipgloss.NewStyle().Width(m.width-12).Height(3).Padding(0, 1)
	buttonContainer := lipgloss.NewStyle().Width(m.width).Height(3).Align(lipgloss.Center, lipgloss.Center)

	return lipgloss.JoinVertical(lipgloss.Top,
		titleBar.Render(titleStyle.Render(m.title)),
		composition.Render(
			inputsContainer.Render(
				lipgloss.JoinVertical(lipgloss.Top,
					strings.Join(formFieldsView, "\n"),
					m.selector.View(),
				),
			),
		),
		buttonContainer.Render(button.Render(m.buttonMsg)),
	)
}

func newUserCreationFormModel() UserFormModel {

	s := selector.New("ReadOnly", "ReadWrite", "Admin")

	s.Prompt = "◇"

	u := UserFormModel{
		inputs:    make([]textinput.Model, 2),
		buttonMsg: "[ SUBMIT ]",
		title:     "Create New User",
		selector:  s,
	}

	for i := range u.inputs {

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

		u.inputs[i] = t

	}

	return u

}

func newUserModificationFormModel() UserFormModel {

	s := selector.New("ReadOnly", "ReadWrite", "Admin")

	s.Prompt = "◇"

	u := UserFormModel{
		inputs:    make([]textinput.Model, 2),
		buttonMsg: "[ CONFIRM ]",
		title:     "Edit User Information",
		selector:  s,
	}

	for i := range u.inputs {

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

		u.inputs[i] = t

	}

	return u

}

// Handles key inputs to navigate and interact with UserFormModel's fields
func (m *UserFormModel) UpdateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up":
			if m.selectedInput > 0 {
				m.selectedInput--
			}
		case "down", "enter":
			if m.selectedInput < len(m.inputs)+1 {
				m.selectedInput++
			}
		}
	}

	for i := range len(m.inputs) {
		if i == m.selectedInput {
			m.inputs[i].Focus()
		} else {
			m.inputs[i].Blur()
		}
	}

	if m.selectedInput == 2 {
		m.selector.Focused = true
	} else {
		m.selector.Focused = false
	}

	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	m.selector.UpdateSelection(msg)

	return tea.Batch(cmds...)
}

func (m *UserFormModel) ClearInputs() {

	for i := range m.inputs {
		m.inputs[i].SetValue("")
	}

}
