package cli

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-git/go-git/v5"

	"github.com/nim-sam/gitport/internal/config"
	"github.com/nim-sam/gitport/internal/database"
	"github.com/nim-sam/gitport/internal/model"
	"github.com/nim-sam/gitport/internal/repository"
	initprompter "github.com/nim-sam/gitport/ui/inline_tui/init_prompter"
)

// initHandler captures the server's starter configuration through the inline
// init prompter, then creates the .gitport directory with the persisted config
// and an initialized database seeded with the admin account.
func initHandler(_ []string) error {
	root, err := os.Getwd()
	if err != nil {
		return err
	}

	if _, err := git.PlainOpen(root); err != nil {
		return fmt.Errorf("not a git repository: %w", err)
	}

	if config.Exists(root) {
		return fmt.Errorf("%s is already initialized; run `gitport start`", config.Dir(root))
	}

	p := tea.NewProgram(initprompter.New(60, 15))

	final, err := p.Run()
	if err != nil {
		return err
	}

	result := final.(initprompter.Model).Result()
	if result == nil {
		fmt.Println("Initialization cancelled. No changes were made.")
		return nil
	}

	cfg := result.GetConfig()
	if err := config.Save(root, cfg); err != nil {
		return err
	}

	if err := database.Init(root); err != nil {
		return err
	}
	repository.Migrate()

	// The wizard always creates exactly one account, which becomes the server
	// Owner regardless of the selected access level, so the configuration is
	// guaranteed to be usable and the Owner role is never assignable through
	// the dashboard.
	admin := result.GetAdmin()
	admin.AccessLevel = model.Owner
	repository.UserRepo.Save(admin)

	fmt.Printf("Initialized %s\n", config.Dir(root))
	fmt.Println("Run `gitport start` to launch the server.")

	return nil
}
