package cli

import (
	"fmt"
	"os"

	"github.com/cjp2600/stepwise/internal/config"
	"github.com/cjp2600/stepwise/internal/logger"
	"github.com/cjp2600/stepwise/internal/workflow"
)

// App represents the CLI application
type App struct {
	config *config.Config
	logger *logger.Logger
}

// NewApp creates a new CLI application
func NewApp(cfg *config.Config, log *logger.Logger) *App {
	return &App{
		config: cfg,
		logger: log,
	}
}

// Run executes the CLI application
func (a *App) Run(args []string) error {
	if len(args) < 2 {
		return a.showHelp()
	}

	command := args[1]

	switch command {
	case "init":
		return a.handleInit(args[2:])
	case "run":
		return a.handleRun(args[2:])
	case "validate":
		return a.handleValidate(args[2:])
	case "info":
		return a.handleInfo(args[2:])
	case "generate":
		return a.handleGenerate(args[2:])
	case "help", "--help", "-h":
		return a.showHelp()
	case "version", "--version", "-v":
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
		return fmt.Errorf("workflow file path is required")
	}

	workflowFile := args[0]
	a.logger.Info("Running workflow", "file", workflowFile)

	// Parse and execute workflow
	wf, err := workflow.Load(workflowFile)
	if err != nil {
		return fmt.Errorf("failed to load workflow: %w", err)
	}

	executor := workflow.NewExecutor(a.config, a.logger)
	results, err := executor.Execute(wf)
	if err != nil {
		return fmt.Errorf("workflow execution failed: %w", err)
	}

	// Print results
	a.printResults(results)
	return nil
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
	help := `Stepwise - API Testing Framework

Usage:
  stepwise <command> [options]

Commands:
  init                    Initialize a new Stepwise project
  run <workflow>         Run a workflow file
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
  --verbose              Enable verbose logging
  --quiet                Enable quiet mode
  --watch                Watch mode for file changes

Examples:
  stepwise init
  stepwise run workflow.yml
  stepwise run workflow.yml --env production
  stepwise validate workflow.yml
  stepwise info workflow.yml

For more information, visit: https://github.com/stepwise/stepwise
`
	fmt.Print(help)
	return nil
}

// showVersion shows the version information
func (a *App) showVersion() error {
	fmt.Println("Stepwise v0.1.0")
	return nil
}

// printResults prints test results
func (a *App) printResults(results []workflow.TestResult) {
	fmt.Println("\nTest Results:")
	fmt.Println("=============")

	passed := 0
	failed := 0
	totalDuration := 0

	for _, result := range results {
		duration := int(result.Duration.Milliseconds())
		totalDuration += duration

		if result.Status == "passed" {
			fmt.Printf("✓ %s (%dms)\n", result.Name, duration)
			passed++
		} else {
			fmt.Printf("✗ %s (%dms) - %s\n", result.Name, duration, result.Error)
			failed++
		}
	}

	fmt.Printf("\nSummary:\n")
	fmt.Printf("- Total: %d tests\n", len(results))
	fmt.Printf("- Passed: %d\n", passed)
	fmt.Printf("- Failed: %d\n", failed)
	fmt.Printf("- Duration: %dms\n", totalDuration)
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
