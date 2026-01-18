package main

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-git/go-git/v5"
)

/*
 * Starts the Terminal User Interface
 */
func StartTui() {
	repo, err := git.PlainOpen(".")
	if err != nil {
		fmt.Println("Not a git repo")
		return
	}

	items, _ := fetchCommits(".", 30)

	// 1. Define fixed dimensions for the inline view
	// Since we are inline, pick a height that fits comfortably
	defaultHeight := 14
	defaultWidth := 80

	l_commit := list.New(items, commitDelegate{listFocused: true}, defaultWidth/2, defaultHeight)
	l_commit.SetShowTitle(false)
	l_commit.SetShowStatusBar(false)

	// 2. Pre-initialize the viewport so 'ready' is true from the start
	viewWidth := defaultWidth - (defaultWidth / 2) - 8
	vp := viewport.New(viewWidth, defaultHeight-2)

	// 3. Populate the initial diff so it's not empty
	var initialHash string
	if len(items) > 0 {
		initialHash = items[0].(CommitItem).hash
		rawDiff := getCommitDiff(repo, initialHash)
		vp.SetContent(highlightDiff(rawDiff))
	}

	cm := commitModel{
		list:         l_commit,
		viewport:     vp,
		repo:         repo,
		ready:        true, // SET THIS TO TRUE
		selectedHash: initialHash,
	}

	// Setup Log Finder
	logItems := []list.Item{
		LogItem{"ERROR", "DB Timeout", "2024-05-20 10:00"},
		LogItem{"INFO", "App Started", "2024-05-20 10:01"},
		LogItem{"WARN", "Disk Near Full", "2024-05-20 10:05"},
	}
	l_log := list.New(logItems, logDelegate{}, 80, 16)
	l_log.SetShowTitle(false)
	l_log.SetShowStatusBar(false)
	l_log.KeyMap.Quit.SetEnabled(false) // Don't let 'q' kill the whole app

	lf := logModel{
		list:  l_log,
		ready: true,
	}

	m := mainModel{
		activeTab: 0, // Start on Commit History to test
		commitLog: cm,
		logFinder: lf,
	}

	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Println("Error:", err)
	}
}

// Entry point (for testing purposes)
func main() {
	StartTui()
}
