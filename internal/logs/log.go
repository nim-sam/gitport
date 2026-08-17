package logs

// Enum categorizing logs into three levels : Warning, Error, Info
type LogLevel int

const (
	Warning LogLevel = iota
	Error
	Info
)

// Struct holding server log information
type Log struct {
	level   LogLevel
	date    string
	message string
}

// Log constructor
func NewLog() *Log {
	return &Log{}
}

// Log getters
func (l Log) Level() LogLevel { return l.level }
func (l Log) Date() string    { return l.date }
func (l Log) Message() string { return l.message }

// Log setters
func (l *Log) WithLevel(level LogLevel) *Log { l.level = level; return l }
func (l *Log) WithDate(date string) *Log     { l.date = date; return l }
func (l *Log) WithMessage(msg string) *Log   { l.message = msg; return l }
