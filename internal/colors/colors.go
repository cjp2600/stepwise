package colors

import (
	"fmt"
	"os"
	"strings"
)

type Colors struct {
	enabled bool
}

func NewColors() *Colors {
	noColor := os.Getenv("NO_COLOR") != ""
	term := os.Getenv("TERM")
	enabled := !noColor &&
		os.Getenv("CI") == "" &&
		os.Getenv("GITHUB_ACTIONS") == "" &&
		os.Getenv("TRAVIS") == "" &&
		os.Getenv("CIRCLECI") == "" &&
		!strings.Contains(term, "dumb") &&
		term != ""
	return &Colors{enabled: enabled}
}

func (c *Colors) Green(text string) string {
	if !c.enabled {
		return text
	}
	return fmt.Sprintf("\033[32m%s\033[0m", text)
}
func (c *Colors) Red(text string) string {
	if !c.enabled {
		return text
	}
	return fmt.Sprintf("\033[31m%s\033[0m", text)
}
func (c *Colors) Yellow(text string) string {
	if !c.enabled {
		return text
	}
	return fmt.Sprintf("\033[33m%s\033[0m", text)
}
func (c *Colors) Blue(text string) string {
	if !c.enabled {
		return text
	}
	return fmt.Sprintf("\033[34m%s\033[0m", text)
}
func (c *Colors) Cyan(text string) string {
	if !c.enabled {
		return text
	}
	return fmt.Sprintf("\033[36m%s\033[0m", text)
}
func (c *Colors) Magenta(text string) string {
	if !c.enabled {
		return text
	}
	return fmt.Sprintf("\033[35m%s\033[0m", text)
}
func (c *Colors) Bold(text string) string {
	if !c.enabled {
		return text
	}
	return fmt.Sprintf("\033[1m%s\033[0m", text)
}
func (c *Colors) Dim(text string) string {
	if !c.enabled {
		return text
	}
	return fmt.Sprintf("\033[2m%s\033[0m", text)
}

// Spinner methods
func (c *Colors) SpinnerFrame(frame int) string {
	if !c.enabled {
		return ""
	}

	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	return frames[frame%len(frames)]
}

func (c *Colors) SpinnerColor(frame int) string {
	if !c.enabled {
		return ""
	}

	colors := []string{"\033[36m", "\033[35m", "\033[34m", "\033[33m", "\033[32m", "\033[31m"}
	color := colors[frame%len(colors)]
	return fmt.Sprintf("%s%s\033[0m", color, c.SpinnerFrame(frame))
}

func (c *Colors) IsEnabled() bool {
	return c.enabled
}
