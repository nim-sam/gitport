package components

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nim-sam/gitport/internal/model"
	"github.com/nim-sam/gitport/ui"
)

// User List

type UserListModel struct {
	Width     int
	Height    int
	ListModel list.Model
	UserData  []model.User
}

func (u UserListModel) Init() tea.Cmd {
	return nil
}

func (u UserListModel) Update(msg tea.Msg) (UserListModel, tea.Cmd) {
	u.ListModel, _ = u.ListModel.Update(msg)
	return u, nil
}

func (u UserListModel) View() string {
	userListModelView := u.ListModel.View()
	return userListModelView
}

// User Item definition

type userItem struct {
	sshKey      string
	name        string
	email       string
	accessLevel model.AccessLevel
}

// User's list.Item implementation

func (i userItem) FilterValue() string { return i.name }
func (i userItem) Title() string       { return i.name + " <" + i.email + ">" }
func (i userItem) Description() string {

	var accessLevelDisplay string
	switch i.accessLevel {
	case 0:
		accessLevelDisplay = "(Read)"
	case 1:
		accessLevelDisplay = "(Write)"
	case 2:
		accessLevelDisplay = "(Admin)"
	}

	return fmt.Sprintf("%s %s", accessLevelDisplay, i.sshKey)
}

// Updates UserListModel's users and sets ListModel's items
func (m *UserListModel) UpdateUsers(users []model.User) {

	m.UserData = users

	// Converting <users> list to list.Item list
	var userList []list.Item
	for _, user := range users {
		userList = append(userList, userItem{
			sshKey:      user.SshKey,
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

func NewUserList(users []model.User) UserListModel {

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

	ul := UserListModel{
		UserData:  users,
		ListModel: listModel,
	}

	// Injecting user data into list model
	ul.UpdateUsers(users)

	return ul
}
