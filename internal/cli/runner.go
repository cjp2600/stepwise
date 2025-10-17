package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

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
	verbose bool
}

// NewWorkflowRunner creates a new workflow runner
func NewWorkflowRunner(cfg *config.Config, log *logger.Logger) *WorkflowRunner {
	colors := NewColors()
	spinner := NewSpinner(colors, "Initializing...")

	// Check if verbose mode is enabled
	verbose := !log.IsMuted()

	return &WorkflowRunner{
		config:  cfg,
		logger:  log,
		colors:  colors,
		spinner: spinner,
		verbose: verbose,
	}
}

// RunWorkflows runs all workflow files in the given path
func (r *WorkflowRunner) RunWorkflows(path string, parallelism int, recursive bool) error {
	if !r.verbose {
		fmt.Println("Loading workflow...")
	} else {
		r.logger.Info("Searching for workflow files...", "path", path)
	}

	workflowFiles, err := r.findWorkflowFiles(path, recursive)
	if err != nil {
		if r.verbose {
			r.logger.Error("Failed to find workflow files", "error", err)
		} else {
			fmt.Printf("✗ Failed to find workflow files: %v\n", err)
		}
		return fmt.Errorf("failed to find workflow files: %w", err)
	}

	if len(workflowFiles) == 0 {
		if r.verbose {
			r.logger.Info("No workflow files found", "path", path)
		} else {
			fmt.Println("ℹ No workflow files found")
		}
		return nil
	}

	if r.verbose {
		r.logger.Info("Found workflow files", "count", len(workflowFiles), "path", path)
	}

	type wfResult struct {
		file    string
		results []workflow.TestResult
		err     error
	}

	resultsCh := make(chan wfResult, len(workflowFiles))

	if parallelism <= 1 {
		// Sequential (old behavior)
		for i, file := range workflowFiles {
			if r.verbose {
				r.logger.Info("Running workflow", "file", file, "progress", fmt.Sprintf("%d/%d", i+1, len(workflowFiles)))
			}

			wf, err := workflow.Load(file)
			if err != nil {
				if r.verbose {
					r.logger.Error("Failed to load workflow", "file", file, "error", err)
				} else {
					r.spinner.Error(fmt.Sprintf("Failed to load workflow: %s", filepath.Base(file)))
				}
				resultsCh <- wfResult{file: file, err: err}
				continue
			}

			// In verbose mode, we don't use spinner at all
			if !r.verbose {
				// Completely disable all logging during workflow execution
				r.logger.SetMuteMode(true)
			}

			executor := workflow.NewExecutor(r.config, r.logger)

			// Setup live progress reporter if not in verbose mode
			var progressReporter *LiveProgressReporter
			if !r.verbose {
				if len(wf.Steps) > 0 {
					progressReporter = NewLiveProgressReporter(r.colors, len(wf.Steps))
					progressReporter.Start()

					// Set progress callback
					executor.SetProgressCallback(func(stepName string, stepIndex int, totalSteps int, status string, duration time.Duration, validationsPassed int, validationsTotal int, err error) {
						update := ProgressUpdate{
							StepName:          stepName,
							StepIndex:         stepIndex,
							TotalSteps:        totalSteps,
							Status:            status,
							Duration:          duration,
							ValidationCount:   validationsTotal,
							ValidationsPassed: validationsPassed,
						}
						if err != nil {
							update.Error = err.Error()
						}
						progressReporter.Update(update)
					})
				}
			}

			res, err := executor.Execute(wf)

			// Stop and complete progress reporter
			if progressReporter != nil {
				progressReporter.Complete()
			}

			if !r.verbose {
				// Re-enable logging after workflow execution
				r.logger.SetMuteMode(false)
			}

			if err != nil {
				if r.verbose {
					r.logger.Error("Workflow failed", "file", file, "error", err)
				}
			} else {
				if r.verbose {
					r.logger.Info("Workflow completed", "file", file)
				}
			}

			// Print collected logs in the report (only in non-verbose mode)
			if !r.verbose {
				logs := r.logger.GetLogBuffer()
				if len(logs) > 0 {
					fmt.Printf("\n%s %s %s\n",
						r.colors.Cyan("==="),
						r.colors.Bold("WORKFLOW LOGS"),
						r.colors.Cyan("==="))
					for _, log := range logs {
						fmt.Println(log)
					}
					fmt.Println()
				}
			}

			resultsCh <- wfResult{file: file, results: res, err: err}
		}
	} else {
		// Parallel worker pool
		if r.verbose {
			r.logger.Info("Running workflows in parallel", "count", len(workflowFiles), "workers", parallelism)
		} else {
			r.spinner.UpdateMessage(fmt.Sprintf("Running %d workflows in parallel (%d workers)", len(workflowFiles), parallelism))
			r.spinner.Restart()
		}

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
					if r.verbose {
						r.logger.Info("Running workflow", "file", file)
					}
					wf, err := workflow.Load(file)
					if err != nil {
						if r.verbose {
							r.logger.Error("Failed to load workflow", "file", file, "error", err)
						}
						resultsCh <- wfResult{file: file, err: err}

						mu.Lock()
						completed++
						if r.verbose {
							r.logger.Info("Workflow completed", "progress", fmt.Sprintf("%d/%d", completed, len(workflowFiles)))
						} else {
							r.spinner.UpdateMessage(fmt.Sprintf("Running workflows: %d/%d completed", completed, len(workflowFiles)))
						}
						mu.Unlock()
						continue
					}

					// In verbose mode, we don't use spinner at all
					if !r.verbose {
						// Stop spinner before executing workflow - logs will be collected but not printed
						r.spinner.Stop()

						// Completely disable all logging during workflow execution
						r.logger.SetMuteMode(true)
					}

					executor := workflow.NewExecutor(r.config, r.logger)
					res, err := executor.Execute(wf)
					resultsCh <- wfResult{file: file, results: res, err: err}

					if !r.verbose {
						// Re-enable logging after workflow execution
						r.logger.SetMuteMode(false)

						// Print collected logs in the report
						logs := r.logger.GetLogBuffer()
						if len(logs) > 0 {
							fmt.Printf("\n%s %s %s\n",
								r.colors.Cyan("==="),
								r.colors.Bold(fmt.Sprintf("WORKFLOW LOGS (%s)", filepath.Base(file))),
								r.colors.Cyan("==="))
							for _, log := range logs {
								fmt.Println(log)
							}
							fmt.Println()
						}
					}

					mu.Lock()
					completed++
					if r.verbose {
						r.logger.Info("Workflow completed", "progress", fmt.Sprintf("%d/%d", completed, len(workflowFiles)))
					} else {
						r.spinner.UpdateMessage(fmt.Sprintf("Running workflows: %d/%d completed", completed, len(workflowFiles)))
					}
					mu.Unlock()
				}
			}()
		}
		wg.Wait()
		if r.verbose {
			r.logger.Info("All workflows completed", "count", len(workflowFiles))
		} else {
			r.spinner.Success(fmt.Sprintf("All %d workflows completed", len(workflowFiles)))
		}
	}
	close(resultsCh)

	// Process results without spinner
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

	// In verbose mode, print processing completion
	if r.verbose {
		r.logger.Info("Results processed successfully")
	}

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
				// Выводим print-текст отдельным блоком, как валидации
				if result.PrintText != "" {
					fmt.Printf("  %s\n", r.colors.Dim(result.PrintText))
				}
				for _, v := range result.Validations {
					var valDesc string
					if v.Error != "" {
						valDesc = v.Error
					} else {
						valDesc = fmt.Sprintf("%s: expected %v, got %v", v.Type, v.Expected, v.Actual)
					}
					if v.Passed {
						fmt.Printf("    %s %s\n", r.colors.Green("✓"), r.colors.Dim(valDesc))
					} else {
						fmt.Printf("    %s %s\n", r.colors.Red("✗"), r.colors.Dim(valDesc))
					}
				}
				passed++
			} else {
				fmt.Printf("%s %s (%dms) - %s\n",
					r.colors.Red("✗"),
					r.colors.Bold(result.Name),
					int(result.Duration.Milliseconds()),
					r.colors.Red(result.Error))
				if result.PrintText != "" {
					fmt.Printf("  %s\n", r.colors.Dim(result.PrintText))
				}
				for _, v := range result.Validations {
					var valDesc string
					if v.Error != "" {
						valDesc = v.Error
					} else {
						valDesc = fmt.Sprintf("%s: expected %v, got %v", v.Type, v.Expected, v.Actual)
					}
					if v.Passed {
						fmt.Printf("    %s %s\n", r.colors.Green("✓"), r.colors.Dim(valDesc))
					} else {
						fmt.Printf("    %s %s\n", r.colors.Red("✗"), r.colors.Dim(valDesc))
					}
				}
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
