package cli

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// ProgressUpdate represents a single progress update
type ProgressUpdate struct {
	StepName          string
	StepIndex         int
	TotalSteps        int
	Status            string // "running", "passed", "failed"
	Duration          time.Duration
	Error             string
	ValidationCount   int
	ValidationsPassed int
}

// LiveProgressReporter displays real-time progress of workflow execution
type LiveProgressReporter struct {
	colors       *Colors
	mu           sync.Mutex
	currentStep  string
	totalSteps   int
	currentIndex int
	passed       int
	failed       int
	running      bool
	startTime    time.Time
	lastUpdate   time.Time
	history      []ProgressUpdate
	maxHistory   int
}

// NewLiveProgressReporter creates a new live progress reporter
func NewLiveProgressReporter(colors *Colors, totalSteps int) *LiveProgressReporter {
	return &LiveProgressReporter{
		colors:     colors,
		totalSteps: totalSteps,
		startTime:  time.Now(),
		lastUpdate: time.Now(),
		history:    make([]ProgressUpdate, 0),
		maxHistory: 5, // Keep last 5 steps in history
	}
}

// Start starts the live progress reporter
func (p *LiveProgressReporter) Start() {
	p.mu.Lock()
	p.running = true
	p.mu.Unlock()

	// Print colored header
	fmt.Println()
	fmt.Println(p.colors.Cyan(strings.Repeat("=", 50)))
	fmt.Println(p.colors.Bold(p.colors.Cyan("Starting Workflow Execution")))
	fmt.Println(p.colors.Cyan(strings.Repeat("=", 50)))
}

// Stop stops the live progress reporter
func (p *LiveProgressReporter) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.running = false
}

// Update updates the progress with a new step
func (p *LiveProgressReporter) Update(update ProgressUpdate) {
	// Always use simple text mode for now (ANSI interactive mode has issues in many terminals)
	// TODO: Add proper terminal capability detection
	if update.Status == "running" {
		fmt.Printf("%s %s %s\n", 
			p.colors.Dim(fmt.Sprintf("[%d/%d]", update.StepIndex, update.TotalSteps)),
			p.colors.Yellow("⟳ Running:"),
			p.colors.Bold(update.StepName))
	} else if update.Status == "passed" {
		validationInfo := ""
		if update.ValidationCount > 0 {
			validationInfo = p.colors.Dim(fmt.Sprintf(" [%d/%d validations passed]", update.ValidationsPassed, update.ValidationCount))
		}
		fmt.Printf("%s %s %s %s%s\n", 
			p.colors.Dim(fmt.Sprintf("[%d/%d]", update.StepIndex, update.TotalSteps)),
			p.colors.Green("✓ PASS:"),
			p.colors.Bold(update.StepName),
			p.colors.Dim(fmt.Sprintf("(%dms)", update.Duration.Milliseconds())),
			validationInfo)
	} else if update.Status == "failed" {
		fmt.Printf("%s %s %s %s - %s\n", 
			p.colors.Dim(fmt.Sprintf("[%d/%d]", update.StepIndex, update.TotalSteps)),
			p.colors.Red("✗ FAIL:"),
			p.colors.Bold(update.StepName),
			p.colors.Dim(fmt.Sprintf("(%dms)", update.Duration.Milliseconds())),
			p.colors.Red(update.Error))
	}
	
	// Update internal state for summary
	p.mu.Lock()
	defer p.mu.Unlock()
	
	if update.Status == "passed" {
		p.passed++
	} else if update.Status == "failed" {
		p.failed++
	}
}

// render renders the current progress display
func (p *LiveProgressReporter) render() {
	if !p.running {
		return
	}

	// Clear previous display
	p.clearDisplay()

	// Move cursor to start of display area
	fmt.Print("\033[s") // Save cursor position

	elapsed := time.Since(p.startTime)
	progress := float64(p.currentIndex) / float64(p.totalSteps) * 100

	// Header
	fmt.Printf("\n%s\n", p.colors.Bold(p.colors.Cyan("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")))
	fmt.Printf("%s %s\n", p.colors.Bold("WORKFLOW EXECUTION"), p.colors.Dim(fmt.Sprintf("(%.1fs elapsed)", elapsed.Seconds())))
	fmt.Printf("%s\n", p.colors.Cyan("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))

	// Progress bar
	barWidth := 40
	filled := int(float64(barWidth) * progress / 100)
	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)

	fmt.Printf("\n%s [%s] %.1f%%\n",
		p.colors.Bold("Progress:"),
		p.colors.Cyan(bar),
		progress)

	// Step counter
	fmt.Printf("%s %d/%d steps completed\n",
		p.colors.Dim("Steps:"),
		p.currentIndex,
		p.totalSteps)

	// Statistics
	fmt.Printf("%s %s  %s %s  %s %d\n",
		p.colors.Bold("Results:"),
		p.colors.Green(fmt.Sprintf("✓ %d passed", p.passed)),
		p.colors.Red(fmt.Sprintf("✗ %d failed", p.failed)),
		p.colors.Dim(fmt.Sprintf("⏱ %s", elapsed.Round(time.Millisecond))),
	)

	// Current step
	if p.currentStep != "" {
		fmt.Printf("\n%s\n", p.colors.Bold(p.colors.Yellow("⟳ Current Step:")))
		fmt.Printf("  %s\n", p.colors.Dim(p.truncate(p.currentStep, 60)))
	}

	// Recent history
	if len(p.history) > 0 {
		fmt.Printf("\n%s\n", p.colors.Bold("Recent Steps:"))
		for i := len(p.history) - 1; i >= 0; i-- {
			h := p.history[i]
			statusIcon := p.colors.Green("✓")
			if h.Status == "failed" {
				statusIcon = p.colors.Red("✗")
			}

			stepName := p.truncate(h.StepName, 45)
			duration := fmt.Sprintf("%dms", h.Duration.Milliseconds())

			line := fmt.Sprintf("  %s %s %s",
				statusIcon,
				p.colors.Dim(stepName),
				p.colors.Dim(fmt.Sprintf("(%s)", duration)))

			if h.ValidationCount > 0 {
				line += p.colors.Dim(fmt.Sprintf(" [%d/%d validations]", h.ValidationsPassed, h.ValidationCount))
			}

			fmt.Println(line)
		}
	}

	fmt.Printf("\n%s\n", p.colors.Cyan("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))

	fmt.Print("\033[u") // Restore cursor position
}

// clearDisplay clears the progress display area
func (p *LiveProgressReporter) clearDisplay() {
	// Move cursor up and clear lines
	// We have approximately 15-17 lines to clear (header + progress + stats + current + history + footer)
	linesToClear := 17
	for i := 0; i < linesToClear; i++ {
		fmt.Print("\033[1A") // Move up one line
		fmt.Print("\033[K")  // Clear line
	}
}

// truncate truncates a string to the specified length
func (p *LiveProgressReporter) truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// Complete marks the progress as complete and shows final summary
func (p *LiveProgressReporter) Complete() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.running = false

	// Print colored summary
	elapsed := time.Since(p.startTime)
	total := p.passed + p.failed
	if total > 0 {
		successRate := float64(p.passed) / float64(total) * 100
		
		// Choose color based on success rate
		summaryColor := p.colors.Green
		if successRate < 100 {
			summaryColor = p.colors.Yellow
		}
		if successRate < 80 {
			summaryColor = p.colors.Red
		}
		
		fmt.Println()
		fmt.Println(p.colors.Cyan(strings.Repeat("=", 50)))
		fmt.Printf("%s %s\n", 
			summaryColor("✓"),
			p.colors.Bold(fmt.Sprintf("Workflow Execution Completed in %.2fs", elapsed.Seconds())))
		fmt.Printf("Steps: %s total, %s passed, %s failed (%s success)\n", 
			p.colors.Bold(fmt.Sprintf("%d", total)),
			p.colors.Green(fmt.Sprintf("%d", p.passed)),
			p.colors.Red(fmt.Sprintf("%d", p.failed)),
			summaryColor(fmt.Sprintf("%.1f%%", successRate)))
		fmt.Println(p.colors.Cyan(strings.Repeat("=", 50)))
		fmt.Println()
	}
}
