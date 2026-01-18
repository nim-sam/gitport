package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

var docStyle = lipgloss.NewStyle()

var (
	addStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#6AB547")) // Green
	delStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF1B1C")) // Red
	headerStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFFF")) // Cyan (for @@ lines)
	baseDiffStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#505050")) // Light grey
)

type CommitItem struct {
	hash, desc, user, time string
}

// Getters for item
func (i CommitItem) Hash() string        { return i.hash }
func (i CommitItem) Description() string { return i.desc }
func (i CommitItem) User() string        { return i.user }
func (i CommitItem) Time() string        { return i.time }
func (i CommitItem) FilterValue() string { return i.hash }

type commitModel struct {
	list         list.Model
	viewport     viewport.Model
	repo         *git.Repository
	ready        bool
	focus        bool   // false = List focused, true = Viewport focused
	selectedHash string // Track current commit to avoid diff re-calculation
}

func (m commitModel) Init() tea.Cmd {
	return nil
}

func (m commitModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		// case "ctrl+c":
		// 	return m, tea.Quit

		case "enter":
			// 1. Toggle the focus boolean
			m.focus = !m.focus

			// 2. Sync the delegate's state to match
			// When m.focus is true, listFocused should be false
			m.list.SetDelegate(commitDelegate{listFocused: !m.focus})

			// 3. Return nil to prevent the "enter" key from
			// triggering the list's default "select" behavior
			return m, nil

		case "esc":
			// Always return to the list on Esc
			if m.focus {
				m.focus = false
				m.list.SetDelegate(commitDelegate{listFocused: true})
				return m, nil
			}
		}

	case tea.WindowSizeMsg:
		targetHeight := 16

		// Width logic remains the same
		listWidth := msg.Width/2 - 4
		viewWidth := msg.Width - listWidth - 8

		m.list.SetSize(listWidth, targetHeight)

		// The viewport internal height MUST be targetHeight - 2
		// so it doesn't try to render 16 lines inside a 14-line visible area
		if !m.ready {
			m.viewport = viewport.New(viewWidth, targetHeight-2)
			m.ready = true
		} else {
			m.viewport.Width = viewWidth
			m.viewport.Height = targetHeight - 2
		}
	}

	// --- Component Interaction Logic (Outside the switch) ---
	if !m.focus {
		var listCmd tea.Cmd
		m.list, listCmd = m.list.Update(msg)
		cmds = append(cmds, listCmd)

		if i, ok := m.list.SelectedItem().(CommitItem); ok {
			if i.hash != m.selectedHash {
				m.selectedHash = i.hash
				rawDiff := getCommitDiff(m.repo, i.hash)
				m.viewport.SetContent(highlightDiff(rawDiff))
				m.viewport.GotoTop()
			}
		}
	} else {
		var viewCmd tea.Cmd
		m.viewport, viewCmd = m.viewport.Update(msg)
		cmds = append(cmds, viewCmd)
	}

	return m, tea.Batch(cmds...)
}

func (m commitModel) View() string {
	if !m.ready {
		return "Initializing..."
	}

	activeColor := lipgloss.Color("#5000ff")
	inactiveColor := lipgloss.Color("240")

	var viewBorderCol lipgloss.Color
	if m.focus {
		viewBorderCol = activeColor
	} else {
		viewBorderCol = inactiveColor
	}

	// 1. List Style: Force it to stretch to targetHeight
	// We don't add a border here so it stays clean
	listSide := lipgloss.NewStyle().
		Width(m.list.Width()).
		Padding(0, 1).
		Render(m.list.View())

	// 2. Viewport Style: Total height (including border) must be targetHeight
	// Since the border takes 2 rows (top + bottom), we set Height to targetHeight - 2
	viewportSide := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(viewBorderCol).
		Padding(0, 1).
		Width(m.viewport.Width).
		Render(m.viewport.View())

	// Join them side-by-side.
	// JoinHorizontal(lipgloss.Top) ensures they align at the very first line.
	mainContent := lipgloss.JoinHorizontal(lipgloss.Top, listSide, viewportSide)

	return docStyle.Render(mainContent)
}

type commitDelegate struct {
	listFocused bool
}

func (d commitDelegate) Height() int                               { return 2 }
func (d commitDelegate) Spacing() int                              { return 1 }
func (d commitDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }

func (d commitDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(CommitItem)
	if !ok {
		return
	}

	listWidth := m.Width()
	if listWidth <= 0 {
		listWidth = 30
	}

	availWidth := listWidth - 11 // Adjusted because we removed the border width
	if availWidth < 10 {
		availWidth = 10
	}

	isSelected := index == m.Index()

	// 1. Define the Hash Color Logic
	// If selected, use a bright color (white or your accent purple)
	// If not selected, use the dim grey
	hashColor := lipgloss.Color("#606060") // Default dim grey
	if isSelected {
		if d.listFocused {
			hashColor = lipgloss.Color("#5000ff") // Accent color when list is active
		} else {
			hashColor = lipgloss.Color("#FFFFFF") // White when list is blurred but item is selected
		}
	}

	// 2. Apply the dynamic color to the hashStyle
	hashStyle := lipgloss.NewStyle().
		Foreground(hashColor).
		Bold(isSelected) // Bold the hash to make it pop even more

	descStyle := lipgloss.NewStyle().Width(availWidth)

	// 3. Clean up the base style (Removed the border logic)
	fn := lipgloss.NewStyle().PaddingLeft(2)

	shortHash := i.hash
	if len(shortHash) > 7 {
		shortHash = shortHash[:7]
	}

	userInfo := lipgloss.NewStyle().Foreground(lipgloss.Color("#707070")).Render(i.user)
	timeInfo := lipgloss.NewStyle().Foreground(lipgloss.Color("#505050")).Render("authored " + i.time)

	// Render the line with the newly colored hash
	line1 := lipgloss.JoinHorizontal(lipgloss.Top, hashStyle.Render(shortHash)+"  ", descStyle.Render(i.desc))
	line2 := fmt.Sprintf("%s %s", userInfo, timeInfo)

	fmt.Fprint(w, fn.Render(line1+"\n"+line2))
}
func fetchCommits(repoPath string, limit int) ([]list.Item, error) {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, err
	}

	ref, err := repo.Head()
	if err != nil {
		return nil, err
	}

	cIter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return nil, err
	}

	var items []list.Item
	count := 0
	err = cIter.ForEach(func(c *object.Commit) error {
		if count >= limit {
			return io.EOF // Stop iterating
		}

		// Clean up trailing whitespace but keep the whole message
		msg := strings.TrimSpace(c.Message)

		items = append(items, CommitItem{
			hash: c.Hash.String(),
			desc: msg,
			user: c.Author.Name,
			time: c.Author.When.Format("Jan 02, 2006"),
		})
		count++
		return nil
	})

	if err != nil && err != io.EOF {
		return nil, err
	}
	return items, nil
}

func getCommitDiff(repo *git.Repository, hash string) string {
	h := plumbing.NewHash(hash)
	commit, err := repo.CommitObject(h)
	if err != nil {
		return "Error finding commit"
	}

	parent, err := commit.Parent(0)
	var patch *object.Patch
	if err != nil {
		// This is the first commit, diff against empty tree
		patch, _ = commit.Patch(nil)
	} else {
		patch, _ = parent.Patch(commit)
	}

	if patch == nil {
		return "No changes found."
	}
	return patch.String()
}

func highlightDiff(rawDiff string) string {
	lines := strings.Split(rawDiff, "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			lines[i] = addStyle.Render(line)
		} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
			lines[i] = delStyle.Render(line)
		} else if strings.HasPrefix(line, "@@") {
			lines[i] = headerStyle.Render(line)
		} else {
			// Apply base style to everything else (filenames, context, etc.)
			lines[i] = baseDiffStyle.Render(line)
		}
	}
	return strings.Join(lines, "\n")
}
