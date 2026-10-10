package logger

import "github.com/nim-sam/gitport/internal/model"

// LogMsg delivers a log entry to a program's update loop. The exported Entry
// lets a consumer build the message without reaching into this package.
type LogMsg struct {
	Entry model.Log
}

// GetLog returns the carried log entry.
func (m LogMsg) GetLog() model.Log {
	return m.Entry
}
