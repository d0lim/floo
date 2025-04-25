package log

import (
	"bytes"
	"fmt"
	"log"
	"time"
)

// LogLevel defines the log level constants.
type LogLevel int

const (
	// DebugLevel is for detailed logs used for debugging.
	DebugLevel LogLevel = iota
	// InfoLevel is for informational logs.
	InfoLevel
	// WarnLevel is for warning logs.
	WarnLevel
	// ErrorLevel is for error logs.
	ErrorLevel
)

// ComponentType represents the component type used in logging.
type ComponentType string

const (
	// GatewayComponent represents the gateway component.
	GatewayComponent ComponentType = "Gateway"
	// ProxyComponent represents the proxy component.
	ProxyComponent ComponentType = "Proxy"
	// FilterComponent represents the filter component.
	FilterComponent ComponentType = "Filter"
	// PredicateComponent represents the predicate component.
	PredicateComponent ComponentType = "Predicate"
)

// LogFlags is a struct that defines the log output format.
type LogFlags struct {
	// Include date
	Date bool
	// Include time
	Time bool
	// Include microseconds
	Microseconds bool
	// Use UTC time
	UTC bool
	// Include file name
	File bool
	// Use full file path
	LongFile bool
}

// Logger is the common interface for Floo logging.
type Logger interface {
	Debug(component ComponentType, format string, v ...interface{})
	Info(component ComponentType, format string, v ...interface{})
	Warn(component ComponentType, format string, v ...interface{})
	Error(component ComponentType, format string, v ...interface{})
	Timed(component ComponentType, format string, v ...interface{}) func(result string)
}

// StandardLogger is the standard logging implementation.
type StandardLogger struct {
	logger *log.Logger
	level  LogLevel
}

var (
	// Current log level, default is InfoLevel
	currentLevel = InfoLevel
	// Default logger
	defaultLogger = log.Default()
	// Shared logger instance
	sharedLogger Logger = &StandardLogger{logger: defaultLogger, level: currentLevel}
)

// SetLogLevel sets the current log level.
func SetLogLevel(level LogLevel) {
	currentLevel = level
	if stdLogger, ok := sharedLogger.(*StandardLogger); ok {
		stdLogger.level = level
	}
}

// GetLogLevel returns the current log level.
func GetLogLevel() LogLevel {
	return currentLevel
}

// IsDebugEnabled checks if debug logging is enabled.
func IsDebugEnabled() bool {
	return currentLevel <= DebugLevel
}

// GetLogger returns the shared logger.
func GetLogger() Logger {
	return sharedLogger
}

// Debug outputs a debug level log.
func (l *StandardLogger) Debug(component ComponentType, format string, v ...interface{}) {
	if l.level <= DebugLevel {
		l.logger.Printf("[%s][DEBUG] %s", component, fmt.Sprintf(format, v...))
	}
}

// Info outputs an info level log.
func (l *StandardLogger) Info(component ComponentType, format string, v ...interface{}) {
	if l.level <= InfoLevel {
		l.logger.Printf("[%s][INFO] %s", component, fmt.Sprintf(format, v...))
	}
}

// Warn outputs a warning level log.
func (l *StandardLogger) Warn(component ComponentType, format string, v ...interface{}) {
	if l.level <= WarnLevel {
		l.logger.Printf("[%s][WARN] %s", component, fmt.Sprintf(format, v...))
	}
}

// Error outputs an error level log.
func (l *StandardLogger) Error(component ComponentType, format string, v ...interface{}) {
	if l.level <= ErrorLevel {
		l.logger.Printf("[%s][ERROR] %s", component, fmt.Sprintf(format, v...))
	}
}

// Timed returns a logger function that measures the time taken for a task.
func (l *StandardLogger) Timed(component ComponentType, format string, v ...interface{}) func(result string) {
	if l.level > InfoLevel {
		return func(string) {}
	}

	start := time.Now()
	l.Info(component, format, v...)

	return func(result string) {
		elapsed := time.Since(start)
		l.Info(component, "%s: %s (elapsed time: %s)", fmt.Sprintf(format, v...), result, elapsed)
	}
}

// Buffer is a buffer struct for capturing logs.
type Buffer struct {
	buf bytes.Buffer
}

// NewBuffer creates a new log buffer.
func NewBuffer() *Buffer {
	return &Buffer{}
}

// Write implements the bytes.Buffer Write method.
func (lb *Buffer) Write(p []byte) (n int, err error) {
	return lb.buf.Write(p)
}

// String returns the log buffer content as a string.
func (lb *Buffer) String() string {
	return lb.buf.String()
}

// CaptureLogsToBuffer redirects log output to a buffer.
// The returned function restores the original output when called.
func CaptureLogsToBuffer(buffer *Buffer) func() {
	original := log.Writer()
	log.SetOutput(buffer)
	return func() {
		log.SetOutput(original)
	}
}

// ConfigureLogger configures the log format and output.
func ConfigureLogger(flags LogFlags, prefix string) {
	var logFlags int

	if flags.Date {
		logFlags |= log.Ldate
	}
	if flags.Time {
		logFlags |= log.Ltime
	}
	if flags.Microseconds {
		logFlags |= log.Lmicroseconds
	}
	if flags.UTC {
		logFlags |= log.LUTC
	}
	if flags.File {
		if flags.LongFile {
			logFlags |= log.Llongfile
		} else {
			logFlags |= log.Lshortfile
		}
	}

	log.SetFlags(logFlags)
	log.SetPrefix(prefix)

	// Update default logger
	defaultLogger = log.Default()
	if stdLogger, ok := sharedLogger.(*StandardLogger); ok {
		stdLogger.logger = defaultLogger
	} else {
		sharedLogger = &StandardLogger{logger: defaultLogger, level: currentLevel}
	}
}

// ConfigureDefaultLogger configures the logger with common log format settings.
func ConfigureDefaultLogger() {
	ConfigureLogger(LogFlags{
		Date: true,
		Time: true,
		File: true,
	}, "[FLOO] ")
}
