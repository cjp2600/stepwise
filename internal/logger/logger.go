package logger

import (
	"log"
	"os"
)

// Logger provides logging functionality
type Logger struct {
	*log.Logger
	level string
}

// New creates a new logger instance
func New() *Logger {
	return &Logger{
		Logger: log.New(os.Stdout, "", log.LstdFlags),
		level:  "info",
	}
}

// SetLevel sets the logging level
func (l *Logger) SetLevel(level string) {
	l.level = level
}

// Info logs an info message
func (l *Logger) Info(msg string, args ...interface{}) {
	if l.level == "info" || l.level == "debug" {
		l.Printf("[INFO] "+msg, args...)
	}
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, args ...interface{}) {
	if l.level == "debug" {
		l.Printf("[DEBUG] "+msg, args...)
	}
}

// Error logs an error message
func (l *Logger) Error(msg string, args ...interface{}) {
	l.Printf("[ERROR] "+msg, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, args ...interface{}) {
	l.Printf("[WARN] "+msg, args...)
}
