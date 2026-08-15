package logger

// Enum categorizing logs into three levels : Warning, Error, Info
type LogLevel int

const (
	Warning LogLevel = iota
	Error
	Info
)

// Struct holding server log information
type log struct {
	level   LogLevel
	time    string
	message string
}

// Log constructor
func New() *log {
	return &log{}
}

// Log getters
func (l log) Level() LogLevel { return l.level }
func (l log) Time() string    { return l.time }
func (l log) Message() string { return l.message }

// Log setters
func (l *log) WithLevel(level LogLevel) *log { l.level = level; return l }
func (l *log) WithTime(time string) *log     { l.time = time; return l }
func (l *log) WithMessage(msg string) *log   { l.message = msg }
