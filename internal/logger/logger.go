package logger

import (
	"fmt"
	"log"
	"os"

	"github.com/cjp2600/stepwise/internal/colors"
)

// LogCallback is a function type for handling log messages
type LogCallback func(level, message string)

// Logger provides logging functionality
type Logger struct {
	*log.Logger
	level      string
	colors     *colors.Colors
	callback   LogCallback
	silentMode bool
	logBuffer  []string
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

// SetSilentMode enables or disables silent mode
func (l *Logger) SetSilentMode(silent bool) {
	l.silentMode = silent
	if silent {
		l.logBuffer = make([]string, 0)
	}
}

// IsMuted returns true if the logger is in mute mode
func (l *Logger) IsMuted() bool {
	return l.silentMode
}

// SetMuteMode completely disables all logging
func (l *Logger) SetMuteMode(mute bool) {
	l.silentMode = mute
	if mute {
		l.logBuffer = make([]string, 0)
	}
}

// GetLogBuffer returns collected logs in silent mode
func (l *Logger) GetLogBuffer() []string {
	return l.logBuffer
}

// PrintLogBuffer prints collected logs
func (l *Logger) PrintLogBuffer() {
	if l.silentMode && len(l.logBuffer) > 0 {
		for _, log := range l.logBuffer {
			fmt.Println(log)
		}
	}
}

// Info logs an info message
func (l *Logger) Info(msg string, args ...interface{}) {
	if l.silentMode {
		return // Completely ignore in mute mode
	}

	if l.level == "info" || l.level == "debug" {
		prefix := l.colors.Green("[INFO]")
		message := prefix + " " + msg

		if l.silentMode {
			l.logBuffer = append(l.logBuffer, fmt.Sprintf(message, args...))
		} else {
			l.Printf(message, args...)
		}

		if l.callback != nil {
			l.callback("INFO", message)
		}
	}
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, args ...interface{}) {
	if l.silentMode {
		return // Completely ignore in mute mode
	}

	if l.level == "debug" {
		prefix := l.colors.Blue("[DEBUG]")
		message := prefix + " " + msg

		if l.silentMode {
			l.logBuffer = append(l.logBuffer, fmt.Sprintf(message, args...))
		} else {
			l.Printf(message, args...)
		}

		if l.callback != nil {
			l.callback("DEBUG", message)
		}
	}
}

// Error logs an error message
func (l *Logger) Error(msg string, args ...interface{}) {
	if l.silentMode {
		return // Completely ignore in mute mode
	}

	prefix := l.colors.Red("[ERROR]")
	message := prefix + " " + msg

	if l.silentMode {
		l.logBuffer = append(l.logBuffer, fmt.Sprintf(message, args...))
	} else {
		l.Printf(message, args...)
	}

	if l.callback != nil {
		l.callback("ERROR", message)
	}
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, args ...interface{}) {
	if l.silentMode {
		return // Completely ignore in mute mode
	}

	prefix := l.colors.Yellow("[WARN]")
	message := prefix + " " + msg

	if l.silentMode {
		l.logBuffer = append(l.logBuffer, fmt.Sprintf(message, args...))
	} else {
		l.Printf(message, args...)
	}

	if l.callback != nil {
		l.callback("WARN", message)
	}
}
