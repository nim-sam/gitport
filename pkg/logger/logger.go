package logger

import (
	"os"
	"path/filepath"

	"github.com/charmbracelet/log"
	"github.com/muesli/termenv"
)

const (
	Logs  = "logs.txt"
	Users = "users.json"
	Conf  = "config.json"
)

type ConfigData struct {
	Public      bool   `json:"public"`
	DefaultPerm string `json:"default_perm"`
}

var WorkDir string
var Config ConfigData

type MultiLogger struct {
	FileLogger *log.Logger
	TermLogger *log.Logger
}

var Logger = MultiLogger{
	FileLogger: nil,
	TermLogger: log.New(os.Stdout),
}

func InitTermLogger() {
	// Set global defaults for all loggers (including middleware)
	log.SetFormatter(log.TextFormatter)
	log.SetTimeFormat("2006-01-02 15:04:05")
	log.SetReportTimestamp(true)

	// Apply settings to our terminal logger instance
	Logger.TermLogger.SetFormatter(log.TextFormatter)
	Logger.TermLogger.SetTimeFormat("2006-01-02 15:04:05")
	Logger.TermLogger.SetReportTimestamp(true)
	Logger.TermLogger.SetColorProfile(termenv.TrueColor)
}

func InitFileLogs() *os.File {
	file, err := os.OpenFile(filepath.Join(WorkDir, Logs), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Error("Could not open logs file", "error", err)
		return nil
	}

	Logger.FileLogger = log.New(file)
	Logger.FileLogger.SetFormatter(log.TextFormatter)
	Logger.FileLogger.SetTimeFormat("2006-01-02 15:04:05")
	Logger.FileLogger.SetReportTimestamp(true)

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
