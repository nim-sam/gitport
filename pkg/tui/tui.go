package tui

import (
	"encoding/base64"

	"fmt"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-git/go-git/v5"

	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"

	"github.com/nim-sam/gitport/pkg/auth"
)

/*
 * Middleware function that provides the TUI interface for SSH sessions
 * When users connect without a git command, they get the TUI instead
 */
func Middleware(repoPath string) wish.Middleware {
	return func(next ssh.Handler) ssh.Handler {
		return func(sess ssh.Session) {
			// Git will have a command included, so only run the TUI if there are no
			// commands passed to ssh.
			if len(sess.Command()) != 0 {
				next(sess)
				return
			}

			// Check if user is admin
			pubKey := sess.PublicKey()
			if pubKey == nil {
				wish.Errorln(sess, "Authentication required")
				next(sess)
				return
			}

			userKey := pubKey.Type() + " " + base64.StdEncoding.EncodeToString(pubKey.Marshal())
			user, exists := auth.GetUserByKey(userKey)

			if !exists || user.Perm != "admin" {
				wish.Errorln(sess, "Access denied: Admin permission required to access TUI")
				next(sess)
				return
			}

			// Open the git repository
			repo, err := git.PlainOpen(repoPath)
			if err != nil {
				wish.Errorln(sess, "Error opening repository:", err)
				next(sess)
				return
			}

			items, err := fetchCommits(repoPath, 30)
			if err != nil {
				wish.Errorln(sess, "Error fetching commits:", err)
				next(sess)
				return
			}

			// // Define fixed dimensions for the TUI
			// defaultHeight := 16
			// defaultWidth := 80

			// l := list.New(items, commitDelegate{listFocused: true}, defaultWidth/2, defaultHeight)
			// l.SetShowTitle(false)
			// l.SetShowStatusBar(false)

			// // Pre-initialize the viewport
			// viewWidth := defaultWidth - (defaultWidth / 2) - 8
			// vp := viewport.New(viewWidth, defaultHeight-2)

			// // Populate the initial diff
			// var initialHash string
			// if len(items) > 0 {
			// 	initialHash = items[0].(CommitItem).hash
			// 	rawDiff := getCommitDiff(repo, initialHash)
			// 	vp.SetContent(highlightDiff(rawDiff))
			// }

			// cm := commitModel{
			// 	list:         l,
			// 	viewport:     vp,
			// 	repo:         repo,
			// 	ready:        true,
			// 	selectedHash: initialHash,
			// }

			// // Setup Log Finder
			// logItems := []list.Item{
			// 	LogItem{"ERROR", "DB Timeout", "2024-05-20 10:00"},
			// 	LogItem{"INFO", "App Started", "2024-05-20 10:01"},
			// 	LogItem{"WARN", "Disk Near Full", "2024-05-20 10:05"},
			// }
			// l_log := list.New(logItems, logDelegate{}, 80, 16)
			// l_log.SetShowTitle(false)
			// l_log.SetShowStatusBar(false)
			// l_log.KeyMap.Quit.SetEnabled(false) // Don't let 'q' kill the whole app

			// lf := logModel{
			// 	list:  l_log,
			// 	ready: true,
			// }

			// // Get PTY dimensions
			// pty, _, active := sess.Pty()
			// if !active {
			// 	wish.Errorln(sess, "No PTY requested")
			// 	next(sess)
			// 	return
			// }

			// db := newDashboard()

			// m := mainModel{
			// 	activeTab: 0, // Start on Dashboard
			// 	dashboard: db,
			// 	commitLog: cm,
			// 	logFinder: lf,
			// 	width:     pty.Window.Width,
			// 	height:    pty.Window.Height,
			// }

			// // Create a bubbletea program with the SSH session as input/output
			// p := tea.NewProgram(
			// 	m,
			// 	tea.WithInput(sess),
			// 	tea.WithOutput(sess),
			// 	tea.WithAltScreen(),
			// )

			// if _, err := p.Run(); err != nil {
			// 	wish.Errorln(sess, "Error running TUI:", err)
			// }
			//
			//items, _ := fetchCommits(".", 30)

			// 1. Define fixed dimensions for the inline view
			// Since we are inline, pick a height that fits comfortably

			pty, _, active := sess.Pty()
			if !active {
				wish.Errorln(sess, "No PTY requested")
				return
			}

			// Use these instead of the hardcoded 80 and 14
			w := pty.Window.Width
			h := pty.Window.Height

			contentHeight := h - 2
			if contentHeight < 1 {
				contentHeight = 1
			}

			commitAreaHeight := contentHeight - 3 // Help + viewport border
			if commitAreaHeight < 1 {
				commitAreaHeight = 1
			}

			commitListWidth := w/2 - 2
			if commitListWidth < 10 {
				commitListWidth = 10
			}

			l_commit := list.New(items, commitDelegate{listFocused: true}, commitListWidth, commitAreaHeight)
			l_commit.SetShowTitle(false)
			l_commit.SetShowStatusBar(false)

			// 2. Pre-initialize the viewport so 'ready' is true from the start
			viewWidth := w - commitListWidth - 4
			if viewWidth < 10 {
				viewWidth = 10
			}
			vp := viewport.New(viewWidth, commitAreaHeight)

			// 3. Populate the initial diff so it's not empty
			var initialHash string
			if len(items) > 0 {
				initialHash = items[0].(CommitItem).hash
				rawDiff := getCommitDiff(repo, initialHash)
				vp.SetContent(highlightDiff(rawDiff))
			}

			cm := commitModel{

				list: l_commit,

				viewport:     vp,
				repo:         repo,
				ready:        true, // SET THIS TO TRUE
				selectedHash: initialHash,
			}

			// Setup Log Finder
			logItems := fetchLogItems()    // <--- Use the helper here
			logHeight := contentHeight - 1 // Leave space for footer
			if logHeight < 1 {
				logHeight = 1
			}
			l_log := list.New(logItems, logDelegate{}, w, logHeight)
			l_log.SetShowTitle(false)
			l_log.SetShowStatusBar(false)
			l_log.SetShowPagination(false)
			l_log.KeyMap.Quit.SetEnabled(false) // Don't let 'q' kill the whole app

			lf := logModel{
				list:  l_log,
				ready: true,
			}

			db := newDashboard()

			m := mainModel{
				activeTab: 0, // Start on Commit History to test
				dashboard: db,
				commitLog: cm,
				logFinder: lf,
				width:     w,
				height:    h,
			}

			p := tea.NewProgram(
				m,
				tea.WithInput(sess),  // Directs input from the SSH client
				tea.WithOutput(sess), // Directs output to the SSH client
				tea.WithAltScreen(),
			)
			if _, err := p.Run(); err != nil {
				fmt.Println("Error:", err)
			}

			next(sess)
		}
	}

}
