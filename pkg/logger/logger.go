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
var configMu sync.RWMutex

type MultiLogger struct {
	LogFile    *os.File
	TermLogger *log.Logger
}

var Logger = MultiLogger{
	LogFile:    nil,
	TermLogger: log.New(os.Stdout),
}

var fileWatcher *fsnotify.Watcher
var onUsersChanged func() error

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

// SetUsersReloadCallback sets the callback function to reload users when file changes
func SetUsersReloadCallback(callback func() error) {
	onUsersChanged = callback
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

// InitFileWatcher initializes the file watcher for users.json and config.json
func InitFileWatcher() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create file watcher: %w", err)
	}

	fileWatcher = watcher

	// Add files to watch
	usersPath := filepath.Join(WorkDir, Users)
	configPath := filepath.Join(WorkDir, Conf)

	// Watch the files if they exist
	if _, err := os.Stat(usersPath); err == nil {
		if err := watcher.Add(usersPath); err != nil {
			Logger.Warn("Could not watch users.json", "error", err)
		} else {
			Logger.Info("Started watching file", "file", Users)
		}
	}

	if _, err := os.Stat(configPath); err == nil {
		if err := watcher.Add(configPath); err != nil {
			Logger.Warn("Could not watch config.json", "error", err)
		} else {
			Logger.Info("Started watching file", "file", Conf)
		}
	}

	// Start watching in a goroutine
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

			// Log write operations
			if event.Has(fsnotify.Write) {
				fileName := filepath.Base(event.Name)
				Logger.Info("File modified externally", "file", fileName, "path", event.Name)

				// Reload the modified file
				if fileName == Users {
					// Import cycle issue - we'll need to use a callback
					if onUsersChanged != nil {
						if err := onUsersChanged(); err != nil {
							Logger.Error("Failed to reload users", "error", err)
						}
					}
				} else if fileName == Conf {
					if err := ReloadConfig(); err != nil {
						Logger.Error("Failed to reload config", "error", err)
					}
				}
			}

			// Log rename/remove operations
			if event.Has(fsnotify.Remove) || event.Has(fsnotify.Rename) {
				fileName := filepath.Base(event.Name)
				Logger.Warn("File removed or renamed", "file", fileName, "path", event.Name)

				// Try to re-add the watch after a short delay (file might be recreated)
				go func(path string) {
					time.Sleep(100 * time.Millisecond)
					if _, err := os.Stat(path); err == nil {
						if err := fileWatcher.Add(path); err == nil {
							Logger.Info("Resumed watching file", "file", filepath.Base(path))
						}
					}
				}(event.Name)
			}

		case err, ok := <-fileWatcher.Errors:
			if !ok {
				return
			}
			Logger.Error("File watcher error", "error", err)
		}
	}
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

	file, err := os.Open(filepath.Join(WorkDir, Conf))
	if err != nil {
		return err
	}
	defer file.Close()

	var newConfig ConfigData
	if err := json.NewDecoder(file).Decode(&newConfig); err != nil {
		return err
	}

	SetConfig(newConfig)
	Logger.Info("Config refreshed", "file", Conf, "public", newConfig.Public, "default_perm", newConfig.DefaultPerm)
	return nil
}

// WriteJSONFile writes JSON data to a file with watcher suspension
func WriteJSONFile(filename string, data interface{}) error {
	if WorkDir == "" {
		Logger.Error("WorkDir not set, cannot write file", "file", filename)
		return fmt.Errorf("WorkDir not set")
	}

	filePath := filepath.Join(WorkDir, filename)
	Logger.Info("Writing JSON file", "file", filename, "path", filePath)

	// Temporarily remove from watcher
	if fileWatcher != nil {
		fileWatcher.Remove(filePath)
		defer func() {
			// Re-add to watcher after a short delay
			time.Sleep(50 * time.Millisecond)
			if _, err := os.Stat(filePath); err == nil {
				fileWatcher.Add(filePath)
			}
		}()
	}

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
	filePath := filepath.Join(WorkDir, Logs)

	// Open the file for reading
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	// Initialize CSV reader
	reader := csv.NewReader(file)

	// Read all records at once
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSV: %w", err)
	}

	return records, nil
}
