package logger

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/log"
	"github.com/nim-sam/gitport/internal/database"
	"github.com/nim-sam/gitport/internal/model"
	"github.com/nim-sam/gitport/internal/repository"
)

// Sink consumes log entries. Register one with AddSink.
type Sink func(model.Log)

// Sinks receive every log entry as it is emitted, so the logger can broadcast
// to a live view without its callers holding a reference to that view.
var (
	sinksMu sync.RWMutex
	sinks   []Sink

	out  = make(chan model.Log, 256)
	once sync.Once
)

// AddSink registers a consumer for every entry emitted after the call. The
// first registration starts a single dispatcher goroutine that fans entries
// out; delivery is best-effort, so a stalled sink never blocks logging.
func AddSink(s Sink) {
	sinksMu.Lock()
	sinks = append(sinks, s)
	sinksMu.Unlock()
	once.Do(func() { go dispatch() })
}

func dispatch() {
	for entry := range out {
		sinksMu.RLock()
		for _, s := range sinks {
			s(entry)
		}
		sinksMu.RUnlock()
	}
}

// SetOutput redirects the default charmbracelet logger's writer. The start TUI
// silences it (io.Discard) while it owns the terminal and renders the entries
// broadcast to it instead, so styled output never corrupts the inline view.
func SetOutput(w io.Writer) {
	log.SetOutput(w)
}

// Generates warning log and saves it to database
func Warn(msg string, keyvals ...interface{}) model.Log {
	log.Warn(msg, keyvals...)
	return saveLog(model.Warning, msg, keyvals...)
}

// Generates info log and saves it to database
func Info(msg string, keyvals ...interface{}) model.Log {
	log.Info(msg, keyvals...)
	return saveLog(model.Info, msg, keyvals...)
}

// Generates error log and saves it to database
func Error(msg string, keyvals ...interface{}) model.Log {
	log.Error(msg, keyvals...)
	return saveLog(model.Error, msg, keyvals...)
}

// Creates log struct and saves it to database
func saveLog(level model.LogLevel, msg string, keyvals ...interface{}) model.Log {

	// Fomatting data
	now := time.Now()
	nowFmt := now.Format("2006/01/02 15:04:05")

	var sb strings.Builder
	sb.WriteString(msg)
	sb.WriteString(" ")
	for i, v := range keyvals {
		sb.WriteString(fmt.Sprint(v))

		if i%2 == 0 {
			sb.WriteString("=")
		} else {
			sb.WriteString(" ")
		}
	}

	entry := model.Log{
		Level:   level,
		Date:    nowFmt,
		Message: sb.String(),
	}

	// Persist only once the database is initialized. Commands such as the root
	// help run before database.Init, so logging must degrade gracefully there.
	if database.DB != nil {
		repository.LogRepo.Save(entry)
	}

	// Fan out to any live sinks without blocking the caller: a slow or stalled
	// consumer drops entries rather than stalling the server.
	select {
	case out <- entry:
	default:
	}

	return entry
}
