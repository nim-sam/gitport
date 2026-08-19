package components

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nim-sam/gitport/internal/auth"
	"github.com/nim-sam/gitport/ui"
)

// User List

type UserListModel struct {
	Width     int
	Height    int
	ListModel list.Model
	users     []list.Item
	UserData  []*auth.User
}

func (u UserListModel) Init() tea.Cmd {
	return nil
}

func (u UserListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return u, nil
}

func (u UserListModel) View() string {

	// Title styles
	// u.listModel.Title = "Users"
	// u.listModel.Styles.Title = lipgloss.NewStyle().
	// 	Foreground(palette.accent)

	u.ListModel.SetShowTitle(false)

	// TitleBar styles
	// u.listModel.Styles.TitleBar = lipgloss.NewStyle().
	// 	Border(lipgloss.RoundedBorder(), false, false, true, false).
	// 	BorderForeground(palette.border).
	// 	Width(u.width - 2).
	// 	AlignHorizontal(lipgloss.Center)

	// Other configs
	u.ListModel.SetShowHelp(false)

	u.ListModel.SetStatusBarItemName("user", "users")

	userListModelView := u.ListModel.View()

	return userListModelView

}

// User Item definition

type userItem struct {
	sshKey      string
	name        string
	email       string
	accessLevel auth.AccessLevel
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

func NewUserList(users []*auth.User) UserListModel {

	// Converting <users> list to list.Item list
	var userList []list.Item
	for _, user := range users {
		userList = append(userList, userItem{
			sshKey:      user.SshKey(),
			name:        user.Name(),
			email:       user.Email(),
			accessLevel: user.AccessLevel(),
		})
	}

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

	return UserListModel{
		users:     userList,
		UserData:  users,
		ListModel: list.New(userList, delegate, 60, 14),
	}
}
