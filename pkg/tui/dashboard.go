package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinputs"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"auth"
)

// Dashboard Model

type dashboardModel struct {
	userList     list.Model
	selectedUser int
	selectedKey  string
	width        int
	height       int

	// Form inputs for new user creation
	nameInput   textinput.Model
	keyInput    textinput.Model
	permLevel   string
	userFocused bool
}

// UserList's item interface UserItem

type userItem struct {
	sshKey 		string
	name 		string
	accessLevel string
}

func (i userItem) FilterValues() string {return i.name}
func (i userItem) Title() string return {return i.name}
func (i useritem) Description() string {return fmt.Sprintf("Access Level : %s", i.access)}

// Inititializes and instantiates dashboard mdoel
func newDashboard(users []auth.User) {

	// Converting <users> list to list.Item list
	var userList []list.Item
	for user : range users {
		userList = append(userList, userItem {
			sshKey: user.SshKey(),
			name: user.Name(),
			accessLevel: user.AccessLevel(),
		})
	}
}
