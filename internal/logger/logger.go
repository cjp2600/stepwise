package logger

import (
	"log"
	"os"

	"github.com/cjp2600/stepwise/internal/colors"
)

// Logger provides logging functionality
type Logger struct {
	*log.Logger
	level  string
	colors *colors.Colors
}

// New creates a new logger instance
func New() *Logger {
	return &Logger{
		Logger: log.New(os.Stdout, "", log.LstdFlags),
		level:  "info",
		colors: colors.NewColors(),
	}
}

// SetLevel sets the logging level
func (l *Logger) SetLevel(level string) {
	l.level = level
}

// Info logs an info message
func (l *Logger) Info(msg string, args ...interface{}) {
	if l.level == "info" || l.level == "debug" {
		prefix := l.colors.Green("[INFO]")
		l.Printf(prefix+" "+msg, args...)
	}
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, args ...interface{}) {
	if l.level == "debug" {
		prefix := l.colors.Blue("[DEBUG]")
		l.Printf(prefix+" "+msg, args...)
	}
}

// Error logs an error message
func (l *Logger) Error(msg string, args ...interface{}) {
	prefix := l.colors.Red("[ERROR]")
	l.Printf(prefix+" "+msg, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, args ...interface{}) {
	prefix := l.colors.Yellow("[WARN]")
	l.Printf(prefix+" "+msg, args...)
}
