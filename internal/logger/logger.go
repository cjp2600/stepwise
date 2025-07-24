package logger

import (
	"log"
	"os"

	"github.com/cjp2600/stepwise/internal/colors"
)

// LogCallback is a function type for handling log messages
type LogCallback func(level, message string)

// Logger provides logging functionality
type Logger struct {
	*log.Logger
	level    string
	colors   *colors.Colors
	callback LogCallback
}

// New creates a new logger instance
func New() *Logger {
	return &Logger{
		Logger: log.New(os.Stdout, "", log.LstdFlags),
		level:  "info",
		colors: colors.NewColors(),
	}
}

// SetCallback sets the log callback function
func (l *Logger) SetCallback(callback LogCallback) {
	l.callback = callback
}

// SetLevel sets the logging level
func (l *Logger) SetLevel(level string) {
	l.level = level
}

// Info logs an info message
func (l *Logger) Info(msg string, args ...interface{}) {
	if l.level == "info" || l.level == "debug" {
		prefix := l.colors.Green("[INFO]")
		message := prefix + " " + msg
		l.Printf(message, args...)
		if l.callback != nil {
			l.callback("INFO", message)
		}
	}
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, args ...interface{}) {
	if l.level == "debug" {
		prefix := l.colors.Blue("[DEBUG]")
		message := prefix + " " + msg
		l.Printf(message, args...)
		if l.callback != nil {
			l.callback("DEBUG", message)
		}
	}
}

// Error logs an error message
func (l *Logger) Error(msg string, args ...interface{}) {
	prefix := l.colors.Red("[ERROR]")
	message := prefix + " " + msg
	l.Printf(message, args...)
	if l.callback != nil {
		l.callback("ERROR", message)
	}
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, args ...interface{}) {
	prefix := l.colors.Yellow("[WARN]")
	message := prefix + " " + msg
	l.Printf(message, args...)
	if l.callback != nil {
		l.callback("WARN", message)
	}
}
