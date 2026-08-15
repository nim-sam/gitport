package ui

import (
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/nim-sam/gitport/internal/logger"
)

type ServerLogsModel struct {
	view  viewport.Model
	table table.Model
	logs  []*logger.Log
}
