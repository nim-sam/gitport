package userlist

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nim-sam/gitport/internal/model"
	"github.com/nim-sam/gitport/ui"
)

// User List

type Model struct {
	Width     int
	Height    int
	ListModel list.Model
	UserData  []model.User
}

func (u Model) Init() tea.Cmd {
	return nil
}

func (u Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	u.ListModel, _ = u.ListModel.Update(msg)
	return u, nil
}

func (u Model) View() string {
	ModelView := u.ListModel.View()
	return ModelView
}

// Updates Model's users and sets ListModel's items
func (m *Model) UpdateUsers(users []model.User) {

	m.UserData = users

	// Converting <users> list to list.Item list
	var userList []list.Item
	for _, user := range users {

		var strippedSshKey string
		if len(user.SshKey) > 30 {
			strippedSshKey = user.SshKey[:30]
			strippedSshKey += "…"
		} else {
			strippedSshKey = user.SshKey
		}

		userList = append(userList, userItem{
			id:          user.ID,
			sshKey:      strippedSshKey,
			name:        user.Name,
			email:       user.Email,
			accessLevel: user.AccessLevel,
		})
	}
	m.ListModel.SetItems(userList)

	// Hide status bar if list is empty
	if len(m.UserData) == 0 {
		m.ListModel.SetShowStatusBar(false)
	} else {
		m.ListModel.SetShowStatusBar(true)
	}
}

func New(users []model.User) Model {

	// Defining delegate
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = lipgloss.NewStyle().
		Foreground(ui.Palette.Accent).
		Border(lipgloss.HiddenBorder(), false, false, false, true).
		Padding(0, 1).
		BorderForeground(ui.Palette.Accent).
		Bold(true)
	delegate.Styles.SelectedDesc = lipgloss.NewStyle().
		Foreground(ui.Palette.Accent).
		Border(lipgloss.HiddenBorder(), false, false, false, true).
		Padding(0, 1).
		BorderForeground(ui.Palette.Accent).
		Italic(true)
	delegate.Styles.NormalTitle = lipgloss.NewStyle().
		Foreground(ui.Palette.Fg).
		Border(lipgloss.HiddenBorder(), false, false, false, true).
		Padding(0, 1)
	delegate.Styles.NormalDesc = lipgloss.NewStyle().
		Foreground(ui.Palette.FgMuted).
		Border(lipgloss.HiddenBorder(), false, false, false, true).
		Padding(0, 1).
		Italic(true)
	delegate.Styles.DimmedTitle = lipgloss.NewStyle().
		Foreground(ui.Palette.Fg).
		Border(lipgloss.HiddenBorder(), false, false, false, true).
		Padding(0, 1).
		Faint(true)
	delegate.Styles.DimmedDesc = lipgloss.NewStyle().
		Foreground(ui.Palette.FgMuted).
		Border(lipgloss.HiddenBorder(), false, false, false, true).
		Padding(0, 1).
		Italic(true).
		Faint(true)

	// Creating list model with configs
	listModel := list.New(make([]list.Item, 0), delegate, 60, 14)
	listModel.SetShowHelp(false)
	listModel.SetStatusBarItemName("user", "users")
	listModel.SetFilteringEnabled(true)
	listModel.SetShowTitle(false)

	ul := Model{
		UserData:  users,
		ListModel: listModel,
	}

	// Injecting user data into list model
	ul.UpdateUsers(users)

	return ul
}

// GetSelectedUser returns the user highlighted in the list. It reports false
// when the list is empty. The selected list item is resolved back to a stored
// user by ID, so the result stays correct while a filter is applied (the list's
// index is relative to the filtered view, not to UserData).
func (m Model) GetSelectedUser() (model.User, bool) {
	item, ok := m.ListModel.SelectedItem().(userItem)
	if !ok {
		return model.User{}, false
	}
	for _, u := range m.UserData {
		if u.ID == item.id {
			return u, true
		}
	}
	return model.User{}, false
}
