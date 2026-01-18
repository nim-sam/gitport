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
		case "ctrl+c":
			return m, tea.Quit

		case "enter", "tab":
			if !m.focus {
				m.focus = true
				m.list.SetDelegate(commitDelegate{listFocused: false})
				return m, nil
			}

		case "esc":
			if m.focus {
				m.focus = false
				m.list.SetDelegate(commitDelegate{listFocused: true})
				return m, nil
			}
		}

	case tea.WindowSizeMsg:
		//h, _ := docStyle.GetFrameSize()

		// Set a fixed height for inline mode (e.g., 20 lines)
		// Or use msg.Height if you want it to fill the view without clearing it
		targetHeight := 20

		listWidth := msg.Width / 5
		viewWidth := msg.Width - listWidth - 45

		// Update List
		m.list.SetSize(listWidth, targetHeight)

		// Update Viewport
		if !m.ready {
			m.viewport = viewport.New(viewWidth, targetHeight-2) // -2 for rounded border
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

	// Define colors for focus states
	activeColor := lipgloss.Color("#5000ff")
	inactiveColor := lipgloss.Color("240")

	var viewBorderCol lipgloss.Color
	if m.focus {
		viewBorderCol = activeColor
	} else {
		viewBorderCol = inactiveColor
	}

	listStyle := lipgloss.NewStyle().
		Padding(0, 1)

	// 2. Style the Viewport Panel (The Diff Box)
	// We make it centered by giving it a fixed width and height
	// based on the calculated viewWidth/viewHeight from Update
	viewportStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(viewBorderCol).
		Padding(0, 1)

	listSide := listStyle.Render(m.list.View())

	// We render the viewport inside the box
	viewportSide := viewportStyle.Render(m.viewport.View())

	// Join panels horizontally
	mainContent := lipgloss.JoinHorizontal(lipgloss.Top, listSide, viewportSide)

	return docStyle.Render(mainContent)
}

type commitDelegate struct {
	listFocused bool
}

func (d commitDelegate) Height() int                               { return 3 }
func (d commitDelegate) Spacing() int                              { return 1 }
func (d commitDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }

func (d commitDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(CommitItem)
	if !ok {
		return
	}

	// 1. Safety Check: If m.Width() is 0 (not yet initialized), use a default
	listWidth := m.Width()
	if listWidth <= 0 {
		listWidth = 30 // Fallback width
	}

	// 2. Calculate available space for the description
	// Subtract space for: Border (1), Padding (2), Hash (7), Gap (2)
	// Total subtraction: ~12
	availWidth := listWidth
	// if availWidth < 10 {
	// 	availWidth = 10
	// } // Floor to 10 chars so it doesn't columnize

	isSelected := index == m.Index()
	borderColor := lipgloss.Color("240")
	if isSelected && d.listFocused {
		borderColor = lipgloss.Color("#5000ff")
	}

	// 3. Apply Styles
	hashStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#505050"))

	// Use the calculated width here
	descStyle := lipgloss.NewStyle().Width(availWidth)

	fn := lipgloss.NewStyle().PaddingLeft(2)
	if isSelected {
		fn = fn.Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(borderColor)
	}

	shortHash := i.hash
	if len(shortHash) > 7 {
		shortHash = shortHash[:7]
	}

	userInfo := lipgloss.NewStyle().Foreground(lipgloss.Color("#5000ff")).Render(i.user)
	timeInfo := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("authored " + i.time)

	// Render components
	renderedHash := hashStyle.Render(shortHash)
	renderedDesc := descStyle.Render(i.desc)

	// Join hash and description horizontally so the wrap stays to the right of the hash
	line1 := lipgloss.JoinHorizontal(lipgloss.Top, renderedHash+"  ", renderedDesc)
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

func CommitLog() {
	// items := []list.Item{
	// 	CommitItem{hash: "1231244", desc: "did something", user: "n-samaali", time: "8 hours ago"},
	// 	CommitItem{hash: "5521a2b", desc: "fixed a bug", user: "gopher", time: "2 hours ago"},
	// }

	// // Use the custom delegate here
	// m := commitModel{list: list.New(items, commitDelegate{}, 0, 0)}
	// m.list.Title = "Commit Logs"

	// p := tea.NewProgram(m, tea.WithAltScreen())
	// if _, err := p.Run(); err != nil {
	// 	os.Exit(1)
	// }

	// 1. Fetch the real data
	repo, err := git.PlainOpen(".")
	if err != nil {
		fmt.Println("Not a git repo")
		return
	}

	items, _ := fetchCommits(".", 30)

	l := list.New(items, commitDelegate{listFocused: true}, 0, 0)

	l.Title = "Commit Log"

	// Pass the repo to the model
	m := commitModel{
		list: l,
		repo: repo,
	}

	// REMOVE tea.WithAltScreen() to make it inline
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
	}
}

/*
 * Starts the Terminal User Interface
 */
func StartTui() {
	CommitLog()
}

// Entry point (for testing purposes)
func main() {
	StartTui()
}
