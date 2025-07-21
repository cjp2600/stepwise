package cli

import (
	"fmt"
	"os"
	"strings"

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
	if len(args) < 2 {
		return a.showHelp()
	}

	// Parse flags
	verbose := false
	var command string
	var commandArgs []string

	for i := 1; i < len(args); i++ {
		arg := args[i]

		if strings.HasPrefix(arg, "--") {
			switch arg {
			case "--verbose":
				verbose = true
			case "--help", "-h":
				return a.showHelp()
			case "--version", "-v":
				return a.showVersion()
			default:
				return fmt.Errorf("unknown flag: %s", arg)
			}
		} else if strings.HasPrefix(arg, "-") {
			// Handle short flags
			if strings.Contains(arg, "v") {
				verbose = true
			}
			if strings.Contains(arg, "h") {
				return a.showHelp()
			}
		} else {
			// First non-flag argument is the command
			if command == "" {
				command = arg
			} else {
				commandArgs = append(commandArgs, arg)
			}
		}
	}

	// Set log level based on verbose flag
	if verbose {
		a.logger.SetLevel("debug")
	} else {
		a.logger.SetLevel("info")
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

	path := args[0]

	// Check if path is a directory or specific file
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to access path: %w", err)
	}

	if info.IsDir() {
		// Run all workflows in directory recursively
		runner := NewWorkflowRunner(a.config, a.logger)
		return runner.RunWorkflows(path)
	} else {
		// Run single workflow file
		a.logger.Info("Running workflow", "file", path)

		// Parse and execute workflow
		wf, err := workflow.Load(path)
		if err != nil {
			return fmt.Errorf("failed to load workflow: %w", err)
		}

		executor := workflow.NewExecutor(a.config, a.logger)
		results, err := executor.Execute(wf)
		if err != nil {
			return fmt.Errorf("workflow execution failed: %w", err)
		}

		// Print results and check for failures
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
	help := fmt.Sprintf(`%s - API Testing Framework

Usage:
  stepwise <command> [options]

Commands:
  init                    Initialize a new Stepwise project
  run <path>             Run workflow file or directory (recursive)
  validate <workflow>    Validate a workflow file
  info <workflow>        Show workflow information
  generate               Generate test data
  help                   Show this help message
  version                Show version information

Options:
  --env <environment>    Set environment (default: development)
  --var <key=value>      Set custom variables
  --parallel <n>         Number of parallel executions
  --timeout <duration>   Request timeout
  --output <format>      Output format (console, json, html)
  %s              Enable verbose logging
  --quiet                Enable quiet mode
  --watch                Watch mode for file changes

Examples:
  stepwise init
  stepwise run workflow.yml                    # Run single file
  stepwise run ./examples                     # Run all workflows in directory
  stepwise run .                             # Run all workflows in current directory
  stepwise run ./...                         # Run all workflows recursively
  stepwise run workflow.yml --env production
  stepwise validate workflow.yml
  stepwise info workflow.yml

Recursive Execution:
  When running a directory, Stepwise will:
  - Find all .yml and .yaml files recursively
  - Skip common directories (.git, node_modules, etc.)
  - Execute each workflow file
  - Provide individual and overall summaries

Environment Variables:
  NO_COLOR               Disable colored output
  CI                     Disable colored output (auto-detected)

For more information, visit: https://github.com/stepwise/stepwise
`,
		a.colors.Bold("Stepwise"),
		a.colors.Cyan("--verbose"))
	fmt.Print(help)
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
				a.colors.Bold(result.Name),
				duration)
			passed++
		} else {
			fmt.Printf("%s %s (%dms) - %s\n",
				a.colors.Red("✗"),
				a.colors.Bold(result.Name),
				duration,
				a.colors.Red(result.Error))
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
