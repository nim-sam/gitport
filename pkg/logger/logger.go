package logger

import (
	"os"
	"path/filepath"
	"time"

	"github.com/charmbracelet/log"
	"github.com/muesli/termenv"
)

const (
	Logs  = "logs.json"
	Users = "users.json"
	Conf  = "config.json"
)

var WorkDir string
var Config = make(map[string]string)

type MultiLogger struct {
	FileLogger *log.Logger
	TermLogger *log.Logger
}

var Logger = MultiLogger{
	FileLogger: nil,
	TermLogger: log.New(os.Stdout),
}

func InitTermLogger() {
	Logger.TermLogger.SetFormatter(log.TextFormatter)
	Logger.TermLogger.SetTimeFormat(time.RFC3339)
	Logger.TermLogger.SetColorProfile(termenv.TrueColor)
}

func InitFileLogs() *os.File {
	file, err := os.OpenFile(filepath.Join(WorkDir, Logs), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Error("Could not open logs file", "error", err)
		return nil
	}

	Logger = MultiLogger{
		log.New(file),
		log.New(os.Stdout),
	}

	Logger.FileLogger.SetFormatter(log.JSONFormatter)
	Logger.FileLogger.SetTimeFormat(time.RFC3339)

	return file
}

func (m *MultiLogger) Info(msg interface{}, keyvals ...interface{}) {
	if m.FileLogger != nil {
		m.FileLogger.Info(msg, keyvals...)
	}
	if m.TermLogger != nil {
		m.TermLogger.Info(msg, keyvals...)
	}
}

func (m *MultiLogger) Warn(msg interface{}, keyvals ...interface{}) {
	if m.FileLogger != nil {
		m.FileLogger.Warn(msg, keyvals...)
	}
	if m.TermLogger != nil {
		m.TermLogger.Warn(msg, keyvals...)
	}
}

func (m *MultiLogger) Error(msg interface{}, keyvals ...interface{}) {
	if m.FileLogger != nil {
		m.FileLogger.Error(msg, keyvals...)
	}
	if m.TermLogger != nil {
		m.TermLogger.Error(msg, keyvals...)
	}
}
