package navigator

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	Title       string
	Tabs        []string
	selectedTab int
	size        int
	Styles      Styles
	width       int
	Separator   string
}

// Dimension setter
func (m *Model) Width(width int) { m.width = width }

type Styles struct {
	Title        lipgloss.Style
	TabsBar      lipgloss.Style
	TabFocused   lipgloss.Style
	TabBlurred   lipgloss.Style
	Separator    lipgloss.Style
	TabSeparator lipgloss.Style
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {

	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			m.selectedTab = (m.selectedTab + 1) % m.size
			// Send updated selection via NavigatorMsg
			cmds = append(cmds,
				SendNavigatorMsg(
					NavigatorMsg{
						currentTab: m.selectedTab,
					},
				),
			)
		case "shift+tab":
			m.selectedTab = (m.selectedTab - 1 + len(m.Tabs)) % m.size
			// Send update selection via NavigatorMsg
			cmds = append(cmds,
				SendNavigatorMsg(
					NavigatorMsg{
						currentTab: m.selectedTab,
					},
				),
			)
		}
	}

	return m, tea.Batch(cmds...)
}

// Creates tea.Cmd out of NavigatorMsg `msg`
func SendNavigatorMsg(msg NavigatorMsg) tea.Cmd {
	return func() tea.Msg {
		return msg
	}
}

func (m Model) View() string {

	var tabStyles []string

	for i, _ := range m.Tabs {
		var styledTab string
		if i == m.selectedTab {
			styledTab = m.Styles.TabFocused.Render(m.Tabs[i])
		} else {
			styledTab = m.Styles.TabBlurred.Render(m.Tabs[i])
		}

		tabStyles = append(tabStyles, styledTab)

		if i != len(m.Tabs)-1 {
			tabStyles = append(tabStyles,
				m.Styles.TabSeparator.Render("·"),
			)
		}
	}

	m.Styles.Title = m.Styles.Title.Width(len(m.Title) + 2)
	m.Styles.TabsBar = m.Styles.TabsBar.Width(m.width - len(m.Title) - 2)

	return lipgloss.JoinHorizontal(lipgloss.Left,
		m.Styles.Title.Render(m.Title),
		m.Styles.Separator.Render(m.Separator),
		m.Styles.TabsBar.Render(
			lipgloss.JoinHorizontal(lipgloss.Left, tabStyles...),
		),
	)
}

// Returns new Model with default Styles
func New() Model {

	titleStyle := lipgloss.NewStyle().
		Bold(true)

	tabsStyle := lipgloss.NewStyle().AlignVertical(lipgloss.Center)

	tabFocusedStyle := lipgloss.NewStyle().
		Bold(true).
		Border(lipgloss.ThickBorder(), false, false, true, false).
		AlignHorizontal(lipgloss.Center).
		Padding(0, 1)

	tabBlurredStyle := lipgloss.NewStyle().
		// Italic(true).
		Border(lipgloss.HiddenBorder(), false, false, true, false).
		AlignHorizontal(lipgloss.Center).
		Padding(0, 1)

	separatorStyle := lipgloss.NewStyle().
		Padding(0, 1)

	defaultStyle := Styles{
		Title:      titleStyle,
		TabsBar:    tabsStyle,
		TabFocused: tabFocusedStyle,
		TabBlurred: tabBlurredStyle,
		Separator:  separatorStyle,
	}

	return Model{Styles: defaultStyle, Separator: "|"}
}

func (m *Model) SetTabs(tabs ...string) {
	m.Tabs = tabs
	m.size = len(tabs)
}

func (m *Model) Index() int {
	return m.selectedTab
}
