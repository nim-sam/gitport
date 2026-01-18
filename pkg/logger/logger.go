package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/muesli/termenv"
)

const (
	Logs  = "logs.csv"
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
	LogFile    *os.File
	TermLogger *log.Logger
}

var Logger = MultiLogger{
	LogFile:    nil,
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
	filePath := filepath.Join(WorkDir, Logs)
	
	// Check if file exists to determine if we need to write header
	_, err := os.Stat(filePath)
	fileExists := err == nil
	
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Error("Could not open logs file", "error", err)
		return nil
	}

	Logger.LogFile = file
	
	// Write CSV header if file is new
	if !fileExists {
		_, err = file.WriteString("Date,Time,Level,Message\n")
		if err != nil {
			log.Error("Could not write CSV header", "error", err)
		}
	}

	return file
}

func (m *MultiLogger) writeCSV(level string, msg interface{}, keyvals ...interface{}) {
	if m.LogFile == nil {
		return
	}
	
	now := time.Now()
	date := now.Format("2006-01-02")
	timeStr := now.Format("15:04:05")
	
	// Build message with key-value pairs
	msgStr := fmt.Sprintf("%v", msg)
	if len(keyvals) > 0 {
		msgStr += " "
		for i := 0; i < len(keyvals); i += 2 {
			if i > 0 {
				msgStr += " "
			}
			if i+1 < len(keyvals) {
				msgStr += fmt.Sprintf("%v=%v", keyvals[i], keyvals[i+1])
			} else {
				msgStr += fmt.Sprintf("%v", keyvals[i])
			}
		}
	}
	
	// Escape quotes and commas in message
	msgStr = strings.ReplaceAll(msgStr, "\"", "\"\"")
	if strings.ContainsAny(msgStr, ",\n\"") {
		msgStr = "\"" + msgStr + "\""
	}
	
	line := fmt.Sprintf("%s,%s,%s,%s\n", date, timeStr, level, msgStr)
	m.LogFile.WriteString(line)
}

func (m *MultiLogger) Info(msg interface{}, keyvals ...interface{}) {
	if m.LogFile != nil {
		m.writeCSV("INFO", msg, keyvals...)
	}
	if m.TermLogger != nil {
		m.TermLogger.Info(msg, keyvals...)
	}
}

func (m *MultiLogger) Warn(msg interface{}, keyvals ...interface{}) {
	if m.LogFile != nil {
		m.writeCSV("WARN", msg, keyvals...)
	}
	if m.TermLogger != nil {
		m.TermLogger.Warn(msg, keyvals...)
	}
}

func (m *MultiLogger) Error(msg interface{}, keyvals ...interface{}) {
	if m.LogFile != nil {
		m.writeCSV("ERROR", msg, keyvals...)
	}
	if m.TermLogger != nil {
		m.TermLogger.Error(msg, keyvals...)
	}
}
