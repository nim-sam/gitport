package logger

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
	time    string
	message string
}

// Log constructor
func New() *Log {
	return &Log{}
}

// Log getters
func (l Log) Level() LogLevel { return l.level }
func (l Log) Time() string    { return l.time }
func (l Log) Message() string { return l.message }

// Log setters
func (l *Log) WithLevel(level LogLevel) *Log { l.level = level; return l }
func (l *Log) WithTime(time string) *Log     { l.time = time; return l }
func (l *Log) WithMessage(msg string) *Log   { l.message = msg; return l }
