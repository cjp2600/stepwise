package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/cjp2600/stepwise/internal/config"
	"github.com/cjp2600/stepwise/internal/logger"
	"github.com/cjp2600/stepwise/internal/workflow"
)

// WorkflowRunner handles recursive workflow execution
type WorkflowRunner struct {
	config  *config.Config
	logger  *logger.Logger
	colors  *Colors
	spinner *Spinner
}

// NewWorkflowRunner creates a new workflow runner
func NewWorkflowRunner(cfg *config.Config, log *logger.Logger) *WorkflowRunner {
	colors := NewColors()
	spinner := NewSpinner(colors, "Initializing...")

	// Set up log callback for spinner
	log.SetCallback(func(level, message string) {
		spinner.HandleLog(level, message)
	})

	return &WorkflowRunner{
		config:  cfg,
		logger:  log,
		colors:  colors,
		spinner: spinner,
	}
}

// RunWorkflows runs all workflow files in the given path
func (r *WorkflowRunner) RunWorkflows(path string, parallelism int, recursive bool) error {
	r.spinner.UpdateMessage("Searching for workflow files...")
	r.spinner.Start()

	workflowFiles, err := r.findWorkflowFiles(path, recursive)
	if err != nil {
		r.spinner.Error("Failed to find workflow files")
		return fmt.Errorf("failed to find workflow files: %w", err)
	}

	if len(workflowFiles) == 0 {
		r.spinner.Info("No workflow files found")
		r.logger.Info("No workflow files found", "path", path)
		return nil
	}

	r.spinner.Success(fmt.Sprintf("Found %d workflow files", len(workflowFiles)))
	r.logger.Info("Found workflow files", "count", len(workflowFiles), "path", path)

	type wfResult struct {
		file    string
		results []workflow.TestResult
		err     error
	}

	resultsCh := make(chan wfResult, len(workflowFiles))

	if parallelism <= 1 {
		// Sequential (old behavior)
		for i, file := range workflowFiles {
			r.spinner.UpdateMessage(fmt.Sprintf("Running workflow %d/%d: %s", i+1, len(workflowFiles), filepath.Base(file)))
			r.spinner.Start()

			r.logger.Info("Running workflow", "file", file)
			wf, err := workflow.Load(file)
			if err != nil {
				r.spinner.Error(fmt.Sprintf("Failed to load workflow: %s", filepath.Base(file)))
				r.logger.Error("Failed to load workflow", "file", file, "error", err)
				resultsCh <- wfResult{file: file, err: err}
				continue
			}

			// Stop spinner before executing workflow to avoid conflicts with logs
			r.spinner.Stop()

			executor := workflow.NewExecutor(r.config, r.logger)
			res, err := executor.Execute(wf)

			if err != nil {
				r.spinner.Error(fmt.Sprintf("Workflow failed: %s", filepath.Base(file)))
			} else {
				r.spinner.Success(fmt.Sprintf("Workflow completed: %s", filepath.Base(file)))
			}

			resultsCh <- wfResult{file: file, results: res, err: err}
		}
	} else {
		// Parallel worker pool
		r.spinner.UpdateMessage(fmt.Sprintf("Running %d workflows in parallel (%d workers)", len(workflowFiles), parallelism))
		r.spinner.Start()

		fileCh := make(chan string, len(workflowFiles))
		for _, file := range workflowFiles {
			fileCh <- file
		}
		close(fileCh)

		var wg sync.WaitGroup
		completed := 0
		var mu sync.Mutex

		for i := 0; i < parallelism; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for file := range fileCh {
					r.logger.Info("Running workflow", "file", file)
					wf, err := workflow.Load(file)
					if err != nil {
						r.logger.Error("Failed to load workflow", "file", file, "error", err)
						resultsCh <- wfResult{file: file, err: err}

						mu.Lock()
						completed++
						r.spinner.UpdateMessage(fmt.Sprintf("Running workflows: %d/%d completed", completed, len(workflowFiles)))
						mu.Unlock()
						continue
					}

					// Stop spinner before executing workflow to avoid conflicts with logs
					r.spinner.Stop()

					executor := workflow.NewExecutor(r.config, r.logger)
					res, err := executor.Execute(wf)
					resultsCh <- wfResult{file: file, results: res, err: err}

					mu.Lock()
					completed++
					r.spinner.UpdateMessage(fmt.Sprintf("Running workflows: %d/%d completed", completed, len(workflowFiles)))
					mu.Unlock()
				}
			}()
		}
		wg.Wait()
		r.spinner.Success(fmt.Sprintf("All %d workflows completed", len(workflowFiles)))
	}
	close(resultsCh)

	r.spinner.UpdateMessage("Processing results...")
	r.spinner.Start()

	totalResults := make([]workflow.TestResult, 0)
	totalPassed := 0
	totalFailed := 0
	totalDuration := 0

	for rres := range resultsCh {
		if rres.err != nil {
			totalFailed++
			continue
		}
		r.printWorkflowResults(rres.file, rres.results)
		for _, result := range rres.results {
			totalResults = append(totalResults, result)
			if result.Status == "passed" {
				totalPassed++
			} else {
				totalFailed++
			}
			totalDuration += int(result.Duration.Milliseconds())
		}
	}

	// Stop spinner before printing summary to avoid conflicts with logs
	r.spinner.Stop()
	r.spinner.Success("Results processed successfully")

	r.printSummary(len(workflowFiles), totalPassed, totalFailed, totalDuration)

	if totalFailed > 0 {
		return fmt.Errorf("workflow execution completed with %d failures", totalFailed)
	}

	return nil
}

// findWorkflowFiles finds all workflow files in the given path
func (r *WorkflowRunner) findWorkflowFiles(path string, recursive bool) ([]string, error) {
	var files []string

	if recursive {
		// Recursive search using filepath.Walk
		err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Skip directories that should be ignored
			if info.IsDir() {
				if r.shouldSkipDirectory(filePath) {
					return filepath.SkipDir
				}
				return nil
			}

			// Check if file is a workflow file
			if r.isWorkflowFile(filePath) {
				files = append(files, filePath)
			}

			return nil
		})

		return files, err
	} else {
		// Non-recursive search - only files in the specified directory
		entries, err := os.ReadDir(path)
		if err != nil {
			return nil, err
		}

		for _, entry := range entries {
			if !entry.IsDir() && r.isWorkflowFile(filepath.Join(path, entry.Name())) {
				files = append(files, filepath.Join(path, entry.Name()))
			}
		}

		return files, nil
	}
}

// shouldSkipDirectory checks if a directory should be skipped
func (r *WorkflowRunner) shouldSkipDirectory(path string) bool {
	// Skip common directories that shouldn't contain workflow files
	skipDirs := []string{
		".git",
		"node_modules",
		"vendor",
		"bin",
		"obj",
		"build",
		"dist",
		"target",
		".vscode",
		".idea",
	}

	baseName := filepath.Base(path)
	for _, skipDir := range skipDirs {
		if baseName == skipDir {
			return true
		}
	}

	return false
}

// isWorkflowFile checks if a file is a workflow file
func (r *WorkflowRunner) isWorkflowFile(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	workflowExtensions := []string{".yml", ".yaml"}

	for _, workflowExt := range workflowExtensions {
		if ext == workflowExt {
			return true
		}
	}

	return false
}

// printWorkflowResults prints results for a single workflow
func (r *WorkflowRunner) printWorkflowResults(filePath string, results []workflow.TestResult) {
	fmt.Printf("\n%s %s %s\n",
		r.colors.Cyan("==="),
		r.colors.Bold(filePath),
		r.colors.Cyan("==="))

	passed := 0
	failed := 0
	duration := 0

	for _, result := range results {
		duration += int(result.Duration.Milliseconds())

		// Handle repeat results
		if result.RepeatCount > 0 {
			r.printRepeatResults(result)
			if result.Status == "passed" {
				passed++
			} else {
				failed++
			}
		} else {
			// Regular step result
			if result.Status == "passed" {
				fmt.Printf("%s %s (%dms)\n",
					r.colors.Green("✓"),
					r.colors.Bold(result.Name),
					int(result.Duration.Milliseconds()))
				passed++
			} else {
				fmt.Printf("%s %s (%dms) - %s\n",
					r.colors.Red("✗"),
					r.colors.Bold(result.Name),
					int(result.Duration.Milliseconds()),
					r.colors.Red(result.Error))
				failed++
			}
		}
	}

	fmt.Printf("  %s: %s passed, %s failed, %dms\n",
		r.colors.Dim("Summary"),
		r.colors.Green(fmt.Sprintf("%d", passed)),
		r.colors.Red(fmt.Sprintf("%d", failed)),
		duration)
}

// printRepeatResults prints results for a repeated step
func (r *WorkflowRunner) printRepeatResults(result workflow.TestResult) {
	fmt.Printf("%s %s (repeat: %d iterations)\n",
		r.colors.Blue("🔄"),
		r.colors.Bold(result.Name),
		result.RepeatCount)

	repeatPassed := 0
	repeatFailed := 0

	for i, repeatResult := range result.RepeatResults {
		statusIcon := r.colors.Green("✓")
		if repeatResult.Status != "passed" {
			statusIcon = r.colors.Red("✗")
			repeatFailed++
		} else {
			repeatPassed++
		}

		fmt.Printf("  %s %s (iteration %d) (%dms)\n",
			statusIcon,
			r.colors.Dim(repeatResult.Name),
			i+1,
			int(repeatResult.Duration.Milliseconds()))

		if repeatResult.Error != "" {
			fmt.Printf("    %s %s\n",
				r.colors.Red("Error:"),
				r.colors.Red(repeatResult.Error))
		}
	}

	// Print repeat summary
	repeatStatus := r.colors.Green("passed")
	if repeatFailed > 0 {
		repeatStatus = r.colors.Red("failed")
	}

	fmt.Printf("  %s: %s (%d/%d iterations %s)\n",
		r.colors.Dim("Repeat Summary"),
		repeatStatus,
		repeatPassed,
		result.RepeatCount,
		r.colors.Dim("passed"))
}

// printSummary prints the overall summary
func (r *WorkflowRunner) printSummary(workflowCount, passed, failed, duration int) {
	fmt.Printf("\n%s\n", r.colors.Cyan(strings.Repeat("=", 50)))
	fmt.Printf("%s\n", r.colors.Bold("OVERALL SUMMARY"))
	fmt.Printf("%s\n", r.colors.Dim("==============="))
	fmt.Printf("Workflows: %d\n", workflowCount)
	fmt.Printf("Tests Passed: %s\n", r.colors.Green(fmt.Sprintf("%d", passed)))
	fmt.Printf("Tests Failed: %s\n", r.colors.Red(fmt.Sprintf("%d", failed)))
	fmt.Printf("Total Duration: %dms\n", duration)

	successRate := float64(passed) / float64(passed+failed) * 100
	rateColor := r.colors.Green
	if successRate < 80 {
		rateColor = r.colors.Yellow
	}
	if successRate < 50 {
		rateColor = r.colors.Red
	}
	fmt.Printf("Success Rate: %s\n", rateColor(fmt.Sprintf("%.1f%%", successRate)))
}
