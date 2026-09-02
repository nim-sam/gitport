package selector

import (
	"errors"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	values             []string
	length             int
	SelectedItem       int
	lenghtLongestValue int
	Styles             Styles
	Focused            bool
	Prompt             string
}

type Styles struct {
	Carrousel lipgloss.Style
	Prompt    lipgloss.Style
}

// Model Init() interface method
func (m Model) Init() tea.Cmd {
	return nil
}

// Model Update() interface method
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "left", "h":
			if m.Focused {
				m.SelectedItem = (m.SelectedItem + m.length - 1) % m.length
			}
		case "right", "l":
			if m.Focused {
				m.SelectedItem = (m.SelectedItem + 1) % m.length
			}
		}
	}
	return m, nil
}

// Model View() interface method
func (m Model) View() string {

	m.Styles.Carrousel = m.Styles.Carrousel.
		Width(m.lenghtLongestValue + 2).
		AlignHorizontal(lipgloss.Center)

	arrowStyle := m.Styles.Carrousel.Width(0).AlignHorizontal(lipgloss.Left)

	return lipgloss.JoinHorizontal(lipgloss.Left,
		m.Styles.Prompt.Render(m.Prompt),
		" ",
		arrowStyle.Render("< ")+m.Styles.Carrousel.Render(m.values[m.SelectedItem])+arrowStyle.Render(" >"),
	)
}

func New(fields ...string) Model {

	// Determining length of longest field
	var maxLength int
	for _, field := range fields {
		if len(field) > maxLength {
			maxLength = len(field)
		}
	}

	// defaultStyle := lipgloss.NewStyle()

	return Model{
		values:             fields,
		length:             len(fields),
		lenghtLongestValue: maxLength,
	}
}

// Sets current selected item to
func (m *Model) SetValue(index int) error {
	if index < 0 || index > m.length-1 {
		return errors.New("Out of bounds")
	}
	m.SelectedItem = index
	return nil
}
