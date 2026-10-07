package logger

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/nim-sam/gitport/internal/model"
	"github.com/nim-sam/gitport/internal/repository"
)

func Warn(msg string, keyvals ...interface{}) {
	log.Warn(msg, keyvals...)
	saveLog(model.Warning, msg, keyvals...)
}

func Info(msg string, keyvals ...interface{}) {
	log.Info(msg, keyvals...)
	saveLog(model.Warning, msg, keyvals...)
}

func Error(msg string, keyvals ...interface{}) {
	log.Error(msg, keyvals...)
	saveLog(model.Warning, msg, keyvals...)
}

func saveLog(level model.LogLevel, msg string, keyvals ...interface{}) {

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

	msgFmt := sb.String()

	// Saving logs to database
	log := model.Log{
		Level:   level,
		Date:    nowFmt,
		Message: msgFmt,
	}

	repository.LogRepo.Save(log)
}
