package cli

import (
	"fmt"
	"os"
	"strings"
	"time"

	flag "github.com/spf13/pflag"

	"github.com/cjp2600/stepwise/internal/config"
	"github.com/cjp2600/stepwise/internal/logger"
	"github.com/cjp2600/stepwise/internal/workflow"
)

// App represents the CLI application
type App struct {
	config *config.Config
	logger *logger.Logger
	colors *Colors
}

// NewApp creates a new CLI application
func NewApp(cfg *config.Config, log *logger.Logger) *App {
	return &App{
		config: cfg,
		logger: log,
		colors: NewColors(),
	}
}

// Run executes the CLI application
func (a *App) Run(args []string) error {
	a.logger.SetMuteMode(true)

	if len(args) < 2 {
		return a.showHelp()
	}

	// Find the command (first non-flag argument)
	command := ""
	var commandArgs []string
	for i := 1; i < len(args); i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, "-") && command == "" {
			command = arg
			// All subsequent args (including flags) go to the handler
			commandArgs = args[i+1:]
			break
		}
		// Global help/version
		if arg == "--help" || arg == "-h" {
			return a.showHelp()
		}
		if arg == "--version" || arg == "-v" {
			return a.showVersion()
		}
	}

	if command == "" {
		return a.showHelp()
	}

	switch command {
	case "init":
		return a.handleInit(commandArgs)
	case "run":
		return a.handleRun(commandArgs)
	case "validate":
		return a.handleValidate(commandArgs)
	case "info":
		return a.handleInfo(commandArgs)
	case "generate":
		return a.handleGenerate(commandArgs)
	case "help":
		return a.showHelp()
	case "version":
		return a.showVersion()
	default:
		return fmt.Errorf("unknown command: %s", command)
	}
}

// handleInit handles the init command
func (a *App) handleInit(args []string) error {
	a.logger.Info("Initializing new Stepwise project")

	// Create default workflow file
	workflowContent := `name: "Example Workflow"
version: "1.0"
description: "A sample workflow for Stepwise"

variables:
  base_url: "https://api.example.com"

steps:
  - name: "Health Check"
    request:
      method: "GET"
      url: "{{base_url}}/health"
    validate:
      - status: 200
      - time: "< 1000ms"
`

	if err := os.WriteFile("workflow.yml", []byte(workflowContent), 0644); err != nil {
		return fmt.Errorf("failed to create workflow file: %w", err)
	}

	a.logger.Info("Created workflow.yml")
	return nil
}

// handleRun handles the run command
func (a *App) handleRun(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("workflow file path or directory is required")
	}

	// Use pflag for flexible GNU-style flag parsing
	fs := flag.NewFlagSet("run", flag.ContinueOnError)
	parallelism := fs.IntP("parallel", "p", 1, "Number of parallel workflow executions")
	recursive := fs.BoolP("recursive", "r", false, "Search recursively in subdirectories")
	verbose := fs.BoolP("verbose", "v", false, "Enable verbose logging")
	_ = fs.Parse(args)

	// Find the first non-flag argument as the path
	path := ""
	for _, arg := range fs.Args() {
		if !strings.HasPrefix(arg, "-") {
			path = arg
			break
		}
	}
	if path == "" {
		fmt.Printf("%s %s\n", a.colors.Red("[ERROR]"), a.colors.Red("workflow file path or directory is required"))
		os.Exit(1)
	}

	info, err := os.Stat(path)
	if err != nil {
		fmt.Printf("%s %s: %v\n", a.colors.Red("[ERROR]"), a.colors.Red("Failed to access path"), err)
		os.Exit(1)
	}

	// Set verbose mode
	if *verbose {
		a.logger.SetMuteMode(false)
		a.logger.SetLevel("debug")
	} else {
		a.logger.SetMuteMode(true)
	}

	if info.IsDir() {
		runner := NewWorkflowRunner(a.config, a.logger)
		return runner.RunWorkflows(path, *parallelism, *recursive)
	} else {
		spinner := NewSpinner(a.colors, "Loading workflow...")

		// Only use spinner in non-verbose mode
		if !*verbose {
			spinner.Start()
		}

		a.logger.Info("Running workflow", "file", path)
		wf, err := workflow.Load(path)
		if err != nil {
			if !*verbose {
				spinner.Error("Failed to load workflow")
			}
			return fmt.Errorf("failed to load workflow: %w", err)
		}

		executor := workflow.NewExecutor(a.config, a.logger)

		// Setup live progress reporter if not in verbose mode
		var progressReporter *LiveProgressReporter
		if !*verbose {
			spinner.Stop()
			if len(wf.Steps) > 0 {
				progressReporter = NewLiveProgressReporter(a.colors, len(wf.Steps))
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

		results, err := executor.Execute(wf)

		// Stop and complete progress reporter
		if progressReporter != nil {
			progressReporter.Complete()
		}

		if err != nil {
			if !*verbose {
				fmt.Printf("✗ Workflow execution failed: %v\n", err)
			}
			return fmt.Errorf("workflow execution failed: %w", err)
		}

		// Workflow completion is now shown by progress reporter
		hasFailures := a.printResults(results)
		if hasFailures {
			return fmt.Errorf("workflow execution completed with failures")
		}
		return nil
	}
}

// handleValidate handles the validate command
func (a *App) handleValidate(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("workflow file path is required")
	}

	workflowFile := args[0]
	a.logger.Info("Validating workflow", "file", workflowFile)

	_, err := workflow.Load(workflowFile)
	if err != nil {
		return fmt.Errorf("workflow validation failed: %w", err)
	}

	a.logger.Info("Workflow is valid")
	return nil
}

// handleInfo handles the info command
func (a *App) handleInfo(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("workflow file path is required")
	}

	workflowFile := args[0]
	a.logger.Info("Getting workflow info", "file", workflowFile)

	wf, err := workflow.Load(workflowFile)
	if err != nil {
		return fmt.Errorf("failed to load workflow: %w", err)
	}

	a.printWorkflowInfo(wf)
	return nil
}

// handleGenerate handles the generate command
func (a *App) handleGenerate(args []string) error {
	a.logger.Info("Generating test data")
	// TODO: Implement data generation
	return fmt.Errorf("data generation not implemented yet")
}

// showHelp shows the help message
func (a *App) showHelp() error {
	fmt.Printf("%s\n", a.colors.Bold("Stepwise - API Testing Framework"))
	fmt.Printf("%s\n\n", a.colors.Dim("A powerful tool for testing APIs with YAML-based workflows"))

	fmt.Printf("%s\n", a.colors.Bold("USAGE:"))
	fmt.Printf("  stepwise <command> [options] [arguments]\n\n")

	fmt.Printf("%s\n", a.colors.Bold("COMMANDS:"))
	fmt.Printf("  %s    %s\n", a.colors.Green("run"), "Execute workflow files")
	fmt.Printf("  %s    %s\n", a.colors.Green("validate"), "Validate workflow files")
	fmt.Printf("  %s    %s\n", a.colors.Green("info"), "Show workflow information")
	fmt.Printf("  %s    %s\n", a.colors.Green("init"), "Initialize a new project")
	fmt.Printf("  %s    %s\n", a.colors.Green("generate"), "Generate test data")
	fmt.Printf("  %s    %s\n", a.colors.Green("help"), "Show this help message")
	fmt.Printf("  %s    %s\n", a.colors.Green("version"), "Show version information")

	fmt.Printf("\n%s\n", a.colors.Bold("EXAMPLES:"))
	fmt.Printf("  %s\n", "stepwise run workflow.yml")
	fmt.Printf("  %s\n", "stepwise run --verbose workflow.yml")
	fmt.Printf("  %s\n", "stepwise run --parallel 4 --recursive ./tests")
	fmt.Printf("  %s\n", "stepwise validate workflow.yml")
	fmt.Printf("  %s\n", "stepwise info workflow.yml")

	fmt.Printf("\n%s\n", a.colors.Bold("GLOBAL OPTIONS:"))
	fmt.Printf("  %s    %s\n", a.colors.Cyan("--help, -h"), "Show help message")
	fmt.Printf("  %s    %s\n", a.colors.Cyan("--version, -v"), "Show version information")

	fmt.Printf("\n%s\n", a.colors.Bold("RUN COMMAND OPTIONS:"))
	fmt.Printf("  %s    %s\n", a.colors.Cyan("--parallel, -p"), "Number of parallel workflow executions (default: 1)")
	fmt.Printf("  %s    %s\n", a.colors.Cyan("--recursive, -r"), "Search recursively in subdirectories")
	fmt.Printf("  %s    %s\n", a.colors.Cyan("--verbose, -v"), "Enable verbose logging (shows detailed logs in real-time)")

	fmt.Printf("\n%s\n", a.colors.Bold("WORKFLOW FILES:"))
	fmt.Printf("  Stepwise supports YAML workflow files with the following features:\n")
	fmt.Printf("  • HTTP and gRPC requests\n")
	fmt.Printf("  • Variable substitution\n")
	fmt.Printf("  • Request validation\n")
	fmt.Printf("  • Response capture\n")
	fmt.Printf("  • Retry logic\n")
	fmt.Printf("  • Parallel execution\n")
	fmt.Printf("  • Component imports\n")

	return nil
}

// showVersion shows the version information
func (a *App) showVersion() error {
	fmt.Println("Stepwise v0.1.0")
	return nil
}

// printResults prints test results and returns true if there were failures
func (a *App) printResults(results []workflow.TestResult) bool {
	fmt.Println("\n" + a.colors.Bold("Test Results:"))
	fmt.Println(a.colors.Dim("============="))

	passed := 0
	failed := 0
	totalDuration := 0

	for _, result := range results {
		duration := int(result.Duration.Milliseconds())
		totalDuration += duration

		if result.Status == "passed" {
			fmt.Printf("%s %s (%dms)\n",
				a.colors.Green("✓"),
				a.colors.Cyan(a.colors.Bold(result.Name)),
				duration)
			if result.PrintText != "" {
				fmt.Printf("  %s\n", a.colors.Dim(result.PrintText))
			}
		} else {
			fmt.Printf("%s %s (%dms) - %s\n",
				a.colors.Red("✗"),
				a.colors.Cyan(a.colors.Bold(result.Name)),
				duration,
				a.colors.Red(result.Error))
			if result.PrintText != "" {
				fmt.Printf("  %s\n", a.colors.Dim(result.PrintText))
			}
		}

		// Print validations for this step
		if len(result.Validations) > 0 {
			fmt.Println("  Validations:")
			for _, v := range result.Validations {
				icon := a.colors.Green("✓")
				lineColor := a.colors.Green
				if !v.Passed {
					icon = a.colors.Red("✗")
					lineColor = a.colors.Red
				}
				msg := fmt.Sprintf("    %s %s: expected %v, got %v", icon, v.Type, v.Expected, v.Actual)
				if v.Error != "" && !v.Passed {
					msg += " (" + v.Error + ")"
				}
				fmt.Println(lineColor(msg))
			}
		}

		// Print repeat results if any
		if result.RepeatCount > 0 && len(result.RepeatResults) > 0 {
			for i, repeatResult := range result.RepeatResults {
				icon := a.colors.Green("✓")
				if repeatResult.Status != "passed" {
					icon = a.colors.Red("✗")
				}
				fmt.Printf("  %s Iteration %d (%dms)\n", icon, i+1, int(repeatResult.Duration.Milliseconds()))
				if repeatResult.Error != "" {
					fmt.Printf("    %s %s\n", a.colors.Red("Error:"), a.colors.Red(repeatResult.Error))
				}
				if len(repeatResult.Validations) > 0 {
					fmt.Println("    Validations:")
					for _, v := range repeatResult.Validations {
						icon := a.colors.Green("✓")
						lineColor := a.colors.Green
						if !v.Passed {
							icon = a.colors.Red("✗")
							lineColor = a.colors.Red
						}
						msg := fmt.Sprintf("      %s %s: expected %v, got %v", icon, v.Type, v.Expected, v.Actual)
						if v.Error != "" && !v.Passed {
							msg += " (" + v.Error + ")"
						}
						fmt.Println(lineColor(msg))
					}
				}
			}
		}
		if result.Status == "passed" {
			passed++
		} else {
			failed++
		}
	}

	fmt.Printf("\n%s\n", a.colors.Bold("Summary:"))
	fmt.Printf("- Total: %d tests\n", len(results))
	fmt.Printf("- Passed: %s\n", a.colors.Green(fmt.Sprintf("%d", passed)))
	fmt.Printf("- Failed: %s\n", a.colors.Red(fmt.Sprintf("%d", failed)))
	fmt.Printf("- Duration: %dms\n", totalDuration)

	return failed > 0
}

// printWorkflowInfo prints workflow information
func (a *App) printWorkflowInfo(wf *workflow.Workflow) {
	fmt.Printf("Workflow: %s\n", wf.Name)
	fmt.Printf("Version: %s\n", wf.Version)
	fmt.Printf("Description: %s\n", wf.Description)
	fmt.Printf("Steps: %d\n", len(wf.Steps))

	if len(wf.Variables) > 0 {
		fmt.Printf("Variables: %d\n", len(wf.Variables))
		for key, value := range wf.Variables {
			fmt.Printf("  %s: %v\n", key, value)
		}
	}
}
