package cli

import (
	"fmt"
	"os"
	"strings"
)

// Colors provides colored output functionality
type Colors struct {
	enabled bool
}

// NewColors creates a new colors instance
func NewColors() *Colors {
	// Check if colors should be disabled
	noColor := os.Getenv("NO_COLOR") != ""
	term := os.Getenv("TERM")

	// Disable colors if:
	// - NO_COLOR environment variable is set
	// - Running in CI (common CI environment variables)
	// - Terminal is not color-capable
	enabled := !noColor &&
		os.Getenv("CI") == "" &&
		os.Getenv("GITHUB_ACTIONS") == "" &&
		os.Getenv("TRAVIS") == "" &&
		os.Getenv("CIRCLECI") == "" &&
		!strings.Contains(term, "dumb") &&
		term != ""

	return &Colors{enabled: enabled}
}

// Green returns green colored text
func (c *Colors) Green(text string) string {
	if !c.enabled {
		return text
	}
	return fmt.Sprintf("\033[32m%s\033[0m", text)
}

// Red returns red colored text
func (c *Colors) Red(text string) string {
	if !c.enabled {
		return text
	}
	return fmt.Sprintf("\033[31m%s\033[0m", text)
}

// Yellow returns yellow colored text
func (c *Colors) Yellow(text string) string {
	if !c.enabled {
		return text
	}
	return fmt.Sprintf("\033[33m%s\033[0m", text)
}

// Blue returns blue colored text
func (c *Colors) Blue(text string) string {
	if !c.enabled {
		return text
	}
	return fmt.Sprintf("\033[34m%s\033[0m", text)
}

// Cyan returns cyan colored text
func (c *Colors) Cyan(text string) string {
	if !c.enabled {
		return text
	}
	return fmt.Sprintf("\033[36m%s\033[0m", text)
}

// Magenta returns magenta colored text
func (c *Colors) Magenta(text string) string {
	if !c.enabled {
		return text
	}
	return fmt.Sprintf("\033[35m%s\033[0m", text)
}

// Bold returns bold text
func (c *Colors) Bold(text string) string {
	if !c.enabled {
		return text
	}
	return fmt.Sprintf("\033[1m%s\033[0m", text)
}

// Dim returns dimmed text
func (c *Colors) Dim(text string) string {
	if !c.enabled {
		return text
	}
	return fmt.Sprintf("\033[2m%s\033[0m", text)
}
