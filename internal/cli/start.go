package cli

import (
	"context"
	"fmt"
	"io"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/nim-sam/gitport/internal/config"
	"github.com/nim-sam/gitport/internal/database"
	"github.com/nim-sam/gitport/internal/logger"
	"github.com/nim-sam/gitport/internal/model"
	"github.com/nim-sam/gitport/internal/repository"
	"github.com/nim-sam/gitport/internal/server"
	serverstatus "github.com/nim-sam/gitport/ui/inline_tui/server_status"
)

// startHandler launches the server under the inline status TUI. The TUI is
// shown first; the bootstrap chain (database, migrations, commit sync, serving)
// then runs in the background, feeding the model status updates as it goes.
func startHandler(_ []string) error {
	root, err := os.Getwd()
	if err != nil {
		return err
	}

	if !config.Exists(root) {
		return fmt.Errorf("gitport is not initialized in %s; run `gitport init` first", root)
	}

	cfg, err := config.Load(root)
	if err != nil {
		return err
	}

	s := server.New(root, cfg)
	s.Ready = make(chan struct{})

	// The status TUI owns the terminal: keep the styled log output off the
	// screen and let the model render the logs broadcast to it instead.
	logger.SetOutput(io.Discard)
	defer logger.SetOutput(os.Stderr)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	info := serverstatus.Info{Name: s.Name, Root: s.Root, Host: s.Host, Port: s.Port}
	statusModel := serverstatus.New(info, 70, 10)

	prog := tea.NewProgram(statusModel)

	// The only place the program is known: forward every log entry the logger
	// emits into the TUI's message loop. No other package needs a program
	// reference to get its logs on screen.
	logger.AddSink(func(entry model.Log) {
		prog.Send(logger.LogMsg{Entry: entry})
	})

	// The TUI is running by the time these goroutines report back: prog.Send
	// blocks until the program's event loop is ready to receive.
	chainDone := make(chan struct{})
	go func() {
		defer close(chainDone)
		runChain(ctx, root, s, prog)
	}()

	_, runErr := prog.Run()

	// Quitting the TUI stops the server; wait for the chain to drain before
	// returning so the listener is fully released.
	cancel()
	<-chainDone

	return runErr
}

// runChain performs the startup sequence the status TUI used to drive, pushing
// state into the UI as it progresses.
func runChain(ctx context.Context, root string, s server.Server, prog *tea.Program) {
	if err := database.Init(root); err != nil {
		prog.Send(serverstatus.StoppedMsg{Err: err})
		return
	}
	repository.Migrate()
	s.InitCommits()

	done := make(chan error, 1)
	go func() {
		done <- s.Run(ctx)
	}()

	select {
	case <-s.Ready:
		prog.Send(serverstatus.RunningMsg{})
	case err := <-done:
		prog.Send(serverstatus.StoppedMsg{Err: err})
		return
	}

	prog.Send(serverstatus.StoppedMsg{Err: <-done})
}
