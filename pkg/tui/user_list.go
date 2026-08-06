package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nim-sam/gitport/pkg/auth"
)

// User List

type UserListModel struct {
	width     int
	height    int
	listModel list.Model
	users     []list.Item
}

func (u UserListModel) Init() tea.Cmd {
	return nil
}

func (u UserListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return u, nil
}

func (u UserListModel) View() string {

	// Title styles
	u.listModel.Title = "User Management"
	u.listModel.Styles.Title = lipgloss.NewStyle().
		Foreground(palette.accent)

	// TitleBar styles
	u.listModel.Styles.TitleBar = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder(), false, false, true, false).
		BorderForeground(palette.border).
		Width(u.width - 2).
		AlignHorizontal(lipgloss.Center)
	u.listModel.SetShowHelp(false)

	userListModelView := u.listModel.View()

	return userListModelView

}

// User Item definition

type userItem struct {
	sshKey      string
	name        string
	accessLevel auth.AccessLevel
}

// User's list.Item implementation

func (i userItem) FilterValue() string { return i.name }
func (i userItem) Title() string       { return i.name }
func (i userItem) Description() string {

	var accessLevelDisplay string
	switch i.accessLevel {
	case 0:
		accessLevelDisplay = "ReadOnly"
	case 1:
		accessLevelDisplay = "ReadWrite"
	case 2:
		accessLevelDisplay = "Admin"
	}

	return fmt.Sprintf("%s ~ %s", accessLevelDisplay, i.sshKey)
}

func newUserList(users []*auth.User) UserListModel {

	// Converting <users> list to list.Item list
	var userList []list.Item
	for _, user := range users {
		userList = append(userList, userItem{
			sshKey:      user.SshKey(),
			name:        user.Name(),
			accessLevel: user.AccessLevel(),
		})
	}

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = lipgloss.NewStyle().
		Foreground(palette.listSelect).
		Border(lipgloss.NormalBorder(), false, false, false, true).
		Padding(0, 1).
		BorderForeground(palette.listSelect)
	delegate.Styles.SelectedDesc = lipgloss.NewStyle().
		Foreground(palette.listSelect).
		Border(lipgloss.NormalBorder(), false, false, false, true).
		Padding(0, 1).
		BorderForeground(palette.listSelect)

	return UserListModel{
		users:     userList,
		listModel: list.New(userList, delegate, 60, 14),
	}
}
