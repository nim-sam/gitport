package logger

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/log"
	"github.com/fsnotify/fsnotify"
)

const (
	Logs  = "logs.csv"
	Users = "users.json"
	Conf  = "config.json"
)

// ConfigData holds the configuration parameters for the server
type ConfigData struct {
	Public      bool   `json:"public"`
	DefaultPerm string `json:"default_perm"`
}

var ConfigDir string
var Config ConfigData
var configMu sync.RWMutex

// sLogger provides file logging capabilities
type sLogger struct {
	LogFile *os.File
	WorkDir string
}

var fileWatcher *fsnotify.Watcher
var onUsersChanged func() error

// Initialize server logger with default terminal logger
var Logger = sLogger{
	LogFile: nil,
	WorkDir: ConfigDir,
}

// InitTermLogger configures the terminal logger with default settings
/*
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
*/

// InitFileLogs initializes file-based logging with CSV format
func (m *sLogger) InitFileLogs(configDir string) *os.File {

	m.WorkDir = configDir

	if m.WorkDir == "" {
		log.Error("WorkDir not set, cannot initialize file logs")
		return nil
	}

	filePath := filepath.Join(m.WorkDir, Logs)
	_, err := os.Stat(filePath)
	fileExists := err == nil

	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Error("Could not open logs file", "error", err)
		return nil
	}

	m.LogFile = file

	// Write CSV header if file is new
	if !fileExists {
		_, err = file.WriteString("Date,Time,Level,Message\n")
		if err != nil {
			log.Error("Could not write CSV header", "error", err)
		}
	}

	return file
}

// SetUsersReloadCallback sets the callback function to reload users when file changes
func SetUsersReloadCallback(callback func() error) {
	onUsersChanged = callback
}

// writeCSV writes a log entry to the CSV file with proper formatting
func (m *sLogger) writeCSV(level string, msg interface{}, keyvals ...interface{}) {
	if m.LogFile == nil {
		return
	}

	now := time.Now()
	date := now.Format("2006-01-02")
	timeStr := now.Format("15:04:05")

	msgStr := m.formatMessage(msg, keyvals...)
	line := fmt.Sprintf("%s,%s,%s,%s\n", date, timeStr, level, msgStr)
	m.LogFile.WriteString(line)
}

// formatMessage formats the message with key-value pairs and proper CSV escaping
func (m *sLogger) formatMessage(msg interface{}, keyvals ...interface{}) string {
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

	// Escape quotes and commas in message for CSV
	msgStr = strings.ReplaceAll(msgStr, "\"", "\"\"")
	if strings.ContainsAny(msgStr, ",\n\"") {
		msgStr = "\"" + msgStr + "\""
	}

	return msgStr
}

// Info logs an informational message
func (m *sLogger) Info(msg interface{}, keyvals ...interface{}) {
	m.log("INFO", msg, keyvals...)
}

// Warn logs a warning message
func (m *sLogger) Warn(msg interface{}, keyvals ...interface{}) {
	m.log("WARN", msg, keyvals...)
}

// Error logs an error message
func (m *sLogger) Error(msg interface{}, keyvals ...interface{}) {
	m.log("ERROR", msg, keyvals...)
}

// log is a helper method to write logs to both file and terminal
func (m *sLogger) log(level string, msg interface{}, keyvals ...interface{}) {
	if m.LogFile != nil {
		m.writeCSV(level, msg, keyvals...)
	}
}

// InitFileWatcher initializes the file watcher for users.json and config.json
func InitFileWatcher() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create file watcher: %w", err)
	}

	fileWatcher = watcher

	// Watch both configuration files
	filesToWatch := []string{Users, Conf}
	for _, filename := range filesToWatch {
		filePath := filepath.Join(ConfigDir, filename)
		if _, err := os.Stat(filePath); err == nil {
			if err := watcher.Add(filePath); err != nil {
				Logger.Warn("Could not watch file", "file", filename, "error", err)
			} else {
				Logger.Info("Started watching file", "file", filename)
			}
		}
	}

	go watchFiles()
	return nil
}

// watchFiles monitors file events and logs changes
func watchFiles() {
	if fileWatcher == nil {
		return
	}

	for {
		select {
		case event, ok := <-fileWatcher.Events:
			if !ok {
				return
			}
			handleFileEvent(event)

		case err, ok := <-fileWatcher.Errors:
			if !ok {
				return
			}
			Logger.Error("File watcher error", "error", err)
		}
	}
}

// handleFileEvent processes file system events
func handleFileEvent(event fsnotify.Event) {
	fileName := filepath.Base(event.Name)

	// Handle write operations
	if event.Has(fsnotify.Write) {
		Logger.Info("File modified externally", "file", fileName, "path", event.Name)
		reloadModifiedFile(fileName, event.Name)
	}

	// Handle rename/remove operations
	if event.Has(fsnotify.Remove) || event.Has(fsnotify.Rename) {
		Logger.Warn("File removed or renamed", "file", fileName, "path", event.Name)
		tryReAddToWatcher(event.Name)
	}
}

// reloadModifiedFile reloads the configuration when a file is modified
func reloadModifiedFile(fileName, filePath string) {
	switch fileName {
	case Users:
		if onUsersChanged != nil {
			if err := onUsersChanged(); err != nil {
				Logger.Error("Failed to reload users", "error", err)
			}
		}
	case Conf:
		if err := ReloadConfig(); err != nil {
			Logger.Error("Failed to reload config", "error", err)
		}
	}
}

// tryReAddToWatcher attempts to re-add a file to the watcher after a delay
func tryReAddToWatcher(filePath string) {
	go func(path string) {
		time.Sleep(100 * time.Millisecond)
		if _, err := os.Stat(path); err == nil {
			if fileWatcher != nil {
				if err := fileWatcher.Add(path); err == nil {
					Logger.Info("Resumed watching file", "file", filepath.Base(path))
				}
			}
		}
	}(filePath)
}

// CloseFileWatcher closes the file watcher
func CloseFileWatcher() {
	if fileWatcher != nil {
		fileWatcher.Close()
	}
}

// GetConfigPublic safely reads the Public config field
func GetConfigPublic() bool {
	configMu.RLock()
	defer configMu.RUnlock()
	return Config.Public
}

// GetConfigDefaultPerm safely reads the DefaultPerm config field
func GetConfigDefaultPerm() string {
	configMu.RLock()
	defer configMu.RUnlock()
	return Config.DefaultPerm
}

// SetConfig safely updates the config with write lock
func SetConfig(newConfig ConfigData) {
	configMu.Lock()
	defer configMu.Unlock()
	Config = newConfig
	Logger.Info("Config updated in memory", "public", newConfig.Public, "default_perm", newConfig.DefaultPerm)
}

// ReloadConfig reloads config from disk (called when file changes)
func ReloadConfig() error {
	Logger.Info("Detected external change, reloading config", "file", Conf)

	file, err := os.Open(filepath.Join(ConfigDir, Conf))
	if err != nil {
		return err
	}
	defer file.Close()

	var newConfig ConfigData
	if err := json.NewDecoder(file).Decode(&newConfig); err != nil {
		return err
	}

	SetConfig(newConfig)
	return nil
}

// WriteJSONFile writes JSON data to a file with watcher suspension
func WriteJSONFile(filename string, data interface{}) error {
	if ConfigDir == "" {
		Logger.Error("ConfigDir not set, cannot write file", "file", filename)
		return fmt.Errorf("ConfigDir not set")
	}

	filePath := filepath.Join(ConfigDir, filename)
	Logger.Info("Writing JSON file", "file", filename, "path", filePath)

	// Temporarily remove from watcher to avoid triggering reload
	suspendFileWatch(filePath)
	defer resumeFileWatch(filePath)

	return writeJSONToFile(filePath, filename, data)
}

// suspendFileWatch temporarily removes a file from the watcher
func suspendFileWatch(filePath string) {
	if fileWatcher != nil {
		fileWatcher.Remove(filePath)
	}
}

// resumeFileWatch re-adds a file to the watcher after a delay
func resumeFileWatch(filePath string) {
	if fileWatcher != nil {
		time.Sleep(50 * time.Millisecond)
		if _, err := os.Stat(filePath); err == nil {
			fileWatcher.Add(filePath)
		}
	}
}

// writeJSONToFile performs the actual JSON file writing
func writeJSONToFile(filePath, filename string, data interface{}) error {
	file, err := os.Create(filePath)
	if err != nil {
		Logger.Error("Failed to create file", "file", filename, "error", err)
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "    ")
	err = encoder.Encode(data)
	if err != nil {
		Logger.Error("Failed to encode JSON", "file", filename, "error", err)
	} else {
		Logger.Info("File written successfully", "file", filename, "path", filePath)
	}
	return err
}

// ReadLogs reads the logs.csv file and returns it as a list of lists.
// Each inner list represents a row: [Date, Time, Level, Message]
func ReadLogs() ([][]string, error) {
	if ConfigDir == "" {
		return nil, fmt.Errorf("ConfigDir not set")
	}

	filePath := filepath.Join(ConfigDir, Logs)
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSV: %w", err)
	}

	return records, nil
}
