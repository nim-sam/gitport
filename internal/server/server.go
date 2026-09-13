package server

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/go-git/go-git/v5"

	// "github.com/nim-sam/gitport/internal/middleware"

	"github.com/nim-sam/gitport/internal/repository"
	"github.com/nim-sam/gitport/internal/util"
	"github.com/nim-sam/gitport/ui/ssh_tui/dashboard"
)

type Server struct {
	Root string
	Name string
	Repo *git.Repository
}
