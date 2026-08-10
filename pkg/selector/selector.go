package selector

import (
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	values             []string
	length             int
	selectedItem       int
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
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

// Helper method updating the the Selector's state based on
// key messages, cycling through its stored values
func (m *Model) UpdateSelection(msg tea.Msg) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "left":
			if m.Focused {
				m.selectedItem = (m.selectedItem + m.length - 1) % m.length
			}
		case "right":
			if m.Focused {
				m.selectedItem = (m.selectedItem + 1) % m.length
			}
		}
	}
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
		arrowStyle.Render("< ")+m.Styles.Carrousel.Render(m.values[m.selectedItem])+arrowStyle.Render(" >"),
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
