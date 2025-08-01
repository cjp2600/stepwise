package cli

import (
	"fmt"
	"sync"
	"time"
)

// Spinner represents a loading spinner
type Spinner struct {
	colors      *Colors
	frame       int
	message     string
	running     bool
	mu          sync.Mutex
	stopChan    chan bool
	doneChan    chan bool
	logHandling bool
}

// NewSpinner creates a new spinner instance
func NewSpinner(colors *Colors, message string) *Spinner {
	return &Spinner{
		colors:      colors,
		message:     message,
		frame:       0,
		running:     false,
		stopChan:    make(chan bool),
		doneChan:    make(chan bool),
		logHandling: true,
	}
}

// Start starts the spinner animation
func (s *Spinner) Start() {
	if !s.colors.IsEnabled() {
		// If colors are disabled, just print the message
		fmt.Print(s.message)
		return
	}

	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.mu.Unlock()

	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-s.stopChan:
				s.doneChan <- true
				return
			case <-ticker.C:
				s.mu.Lock()
				if !s.running {
					s.mu.Unlock()
					return
				}
				s.frame++
				s.mu.Unlock()

				// Clear the line and print the spinner
				fmt.Print("\r")
				fmt.Print(s.colors.SpinnerColor(s.frame))
				fmt.Print(" ")
				fmt.Print(s.message)
			}
		}
	}()
}

// Stop stops the spinner animation
func (s *Spinner) Stop() {
	if !s.colors.IsEnabled() {
		fmt.Println()
		return
	}

	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	s.mu.Unlock()

	s.stopChan <- true
	<-s.doneChan

	// Clear the line and move to next line
	fmt.Print("\r")
	fmt.Print("\033[K") // Clear line
	fmt.Println()       // Move to next line
}

// Disable disables the spinner completely
func (s *Spinner) Disable() {
	s.mu.Lock()
	s.running = false
	s.mu.Unlock()
}

// Restart restarts the spinner with log handling enabled
func (s *Spinner) Restart() {
	s.EnableLogHandling()
	s.Start()
}

// UpdateMessage updates the spinner message
func (s *Spinner) UpdateMessage(message string) {
	s.mu.Lock()
	s.message = message
	s.mu.Unlock()
}

// Success stops the spinner and shows a success message
func (s *Spinner) Success(message string) {
	s.Stop()
	if s.colors.IsEnabled() {
		fmt.Printf("%s %s\n", s.colors.Green("✓"), message)
	} else {
		fmt.Printf("✓ %s\n", message)
	}
}

// Error stops the spinner and shows an error message
func (s *Spinner) Error(message string) {
	s.Stop()
	if s.colors.IsEnabled() {
		fmt.Printf("%s %s\n", s.colors.Red("✗"), message)
	} else {
		fmt.Printf("✗ %s\n", message)
	}
}

// Info stops the spinner and shows an info message
func (s *Spinner) Info(message string) {
	s.Stop()
	if s.colors.IsEnabled() {
		fmt.Printf("%s %s\n", s.colors.Cyan("ℹ"), message)
	} else {
		fmt.Printf("ℹ %s\n", message)
	}
}

// DisableLogHandling disables the spinner's log handling
func (s *Spinner) DisableLogHandling() {
	s.mu.Lock()
	s.logHandling = false
	s.mu.Unlock()
}

// EnableLogHandling enables the spinner's log handling
func (s *Spinner) EnableLogHandling() {
	s.mu.Lock()
	s.logHandling = true
	s.mu.Unlock()
}

// HandleLog handles log messages and adjusts spinner accordingly
func (s *Spinner) HandleLog(level, message string) {
	s.mu.Lock()
	if !s.running || !s.logHandling {
		s.mu.Unlock()
		return
	}
	s.mu.Unlock()

	// Temporarily stop the spinner
	s.mu.Lock()
	s.running = false
	s.mu.Unlock()

	// Wait for the spinner goroutine to stop
	s.stopChan <- true
	<-s.doneChan

	// Clear the current line completely
	fmt.Print("\r")
	fmt.Print("\033[K") // Clear line
	fmt.Println()       // Move to next line

	// Print the log message
	fmt.Println(message)

	// Restart the spinner
	s.mu.Lock()
	s.running = true
	s.mu.Unlock()

	// Start the spinner again
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-s.stopChan:
				s.doneChan <- true
				return
			case <-ticker.C:
				s.mu.Lock()
				if !s.running {
					s.mu.Unlock()
					return
				}
				s.frame++
				s.mu.Unlock()

				// Clear the line and print the spinner
				fmt.Print("\r")
				fmt.Print(s.colors.SpinnerColor(s.frame))
				fmt.Print(" ")
				fmt.Print(s.message)
			}
		}
	}()
}
