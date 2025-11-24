package workflow

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cjp2600/stepwise/internal/config"
	dbclient "github.com/cjp2600/stepwise/internal/database"
	grpcclient "github.com/cjp2600/stepwise/internal/grpc"
	httpclient "github.com/cjp2600/stepwise/internal/http"
	"github.com/cjp2600/stepwise/internal/logger"
	"github.com/cjp2600/stepwise/internal/validation"
	"github.com/cjp2600/stepwise/internal/variables"
	"gopkg.in/yaml.v3"
)

// Import represents an imported component
type Import struct {
	Path      string                 `yaml:"path" json:"path"`
	Version   string                 `yaml:"version,omitempty" json:"version,omitempty"`
	Alias     string                 `yaml:"alias,omitempty" json:"alias,omitempty"`
	Variables map[string]interface{} `yaml:"variables,omitempty" json:"variables,omitempty"`
	Overrides map[string]interface{} `yaml:"overrides,omitempty" json:"overrides,omitempty"`
}

// Workflow represents a complete test workflow
type Workflow struct {
	Name        string                 `yaml:"name" json:"name"`
	Version     string                 `yaml:"version" json:"version"`
	Description string                 `yaml:"description" json:"description"`
	Variables   map[string]interface{} `yaml:"variables" json:"variables"`
	Imports     []Import               `yaml:"imports,omitempty" json:"imports,omitempty"`
	Steps       []Step                 `yaml:"steps" json:"steps"`
	Groups      []StepGroup            `yaml:"groups" json:"groups"`
	Captures    map[string]string      `yaml:"captures,omitempty" json:"captures,omitempty"` // Global captures for the workflow
	SourceFile  string                 `yaml:"-" json:"-"`                                   // путь к исходному workflow-файлу (не сериализуется)
}

// StepGroup represents a group of steps that can be executed together
type StepGroup struct {
	Name        string      `yaml:"name" json:"name"`
	Description string      `yaml:"description" json:"description"`
	Parallel    bool        `yaml:"parallel" json:"parallel"`
	Condition   string      `yaml:"condition" json:"condition"`
	Steps       []Step      `yaml:"steps" json:"steps"`
	Groups      []StepGroup `yaml:"groups,omitempty" json:"groups,omitempty"`
}

// Step represents a single test step
type Step struct {
	Name         string                      `yaml:"name" json:"name"`
	ShowResponse bool                        `yaml:"show_response,omitempty" json:"show_response,omitempty"`
	Use          string                      `yaml:"use,omitempty" json:"use,omitempty"`
	Description  string                      `yaml:"description" json:"description"`
	Request      Request                     `yaml:"request" json:"request"`
	Validate     []validation.ValidationRule `yaml:"validate" json:"validate"`
	Capture      map[string]string           `yaml:"capture" json:"capture"`
	Condition    string                      `yaml:"condition" json:"condition"`
	Retry        int                         `yaml:"retry" json:"retry"`
	RetryDelay   string                      `yaml:"retry_delay" json:"retry_delay"`
	Timeout      string                      `yaml:"timeout" json:"timeout"`
	Repeat       *RepeatConfig               `yaml:"repeat,omitempty" json:"repeat,omitempty"`
	Poll         *PollConfig                 `yaml:"poll,omitempty" json:"poll,omitempty"`           // Polling configuration
	Wait         string                      `yaml:"wait,omitempty" json:"wait,omitempty"`           // Новое поле для задержки
	Print        string                      `yaml:"print,omitempty" json:"print,omitempty"`         // Новое поле для вывода
	Variables    map[string]interface{}      `yaml:"variables,omitempty" json:"variables,omitempty"` // Переменные для переопределения в use
}

// RepeatConfig represents configuration for repeating a step
type RepeatConfig struct {
	Count     int                    `yaml:"count" json:"count"`
	Delay     string                 `yaml:"delay,omitempty" json:"delay,omitempty"`
	Parallel  bool                   `yaml:"parallel,omitempty" json:"parallel,omitempty"`
	Variables map[string]interface{} `yaml:"variables,omitempty" json:"variables,omitempty"`
}

// PollConfig represents configuration for polling a step until condition is met
type PollConfig struct {
	MaxAttempts int                         `yaml:"max_attempts" json:"max_attempts"`             // Maximum number of polling attempts
	Interval    string                      `yaml:"interval,omitempty" json:"interval,omitempty"` // Delay between polling attempts (e.g., "1s", "500ms")
	Until       []validation.ValidationRule `yaml:"until" json:"until"`                           // Conditions that must be met to stop polling
}

// Request represents an HTTP, gRPC, or database request
type Request struct {
	// Protocol type: "http", "grpc", or "db"
	Protocol string `yaml:"protocol" json:"protocol"`

	// HTTP fields
	Method  string            `yaml:"method" json:"method"`
	URL     string            `yaml:"url" json:"url"`
	Headers map[string]string `yaml:"headers" json:"headers"`
	Body    interface{}       `yaml:"body" json:"body"`
	Query   interface{}       `yaml:"query" json:"query"` // map[string]string for HTTP, string for DB
	Auth    *httpclient.Auth  `yaml:"auth" json:"auth"`

	// gRPC fields
	Service    string            `yaml:"service" json:"service"`
	GRPCMethod string            `yaml:"grpc_method" json:"grpc_method"`
	Data       interface{}       `yaml:"data" json:"data"`
	Metadata   map[string]string `yaml:"metadata" json:"metadata"`
	ServerAddr string            `yaml:"server_addr" json:"server_addr"`
	Insecure   bool              `yaml:"insecure" json:"insecure"`

	// Database fields
	DBConfig *dbclient.Config `yaml:"db" json:"db"`

	// Common fields
	Timeout string `yaml:"timeout" json:"timeout"`
}

// TestResult represents the result of a test step
type TestResult struct {
	Name          string                        `json:"name"`
	Status        string                        `json:"status"`
	Duration      time.Duration                 `json:"duration"`
	Error         string                        `json:"error,omitempty"`
	Validations   []validation.ValidationResult `json:"validations,omitempty"`
	CapturedData  map[string]interface{}        `json:"captured_data,omitempty"`
	Retries       int                           `json:"retries,omitempty"`
	RepeatResults []TestResult                  `json:"repeat_results,omitempty"`
	RepeatCount   int                           `json:"repeat_count,omitempty"`
	PrintText     string                        `json:"print_text,omitempty"`    // Текст print для отчёта
	PollAttempts  int                           `json:"poll_attempts,omitempty"` // Количество попыток поллинга
}

// GroupResult represents the result of a step group
type GroupResult struct {
	Name     string        `json:"name"`
	Status   string        `json:"status"`
	Duration time.Duration `json:"duration"`
	Results  []TestResult  `json:"results"`
	Error    string        `json:"error,omitempty"`
	Parallel bool          `json:"parallel"`
}

// StepWithVars связывает шаг с переменными компонента
type StepWithVars struct {
	Step      Step
	Variables map[string]interface{}
}

// Executor executes workflows
// ProgressCallback is called on each step progress update
type ProgressCallback func(stepName string, stepIndex int, totalSteps int, status string, duration time.Duration, validationsPassed int, validationsTotal int, err error)

type Executor struct {
	config           *config.Config
	logger           *logger.Logger
	httpClient       *httpclient.Client
	grpcClient       *grpcclient.Client
	dbClient         *dbclient.Client
	validator        *validation.Validator
	varManager       *variables.Manager
	progressCallback ProgressCallback
	failFast         bool
}

// SetProgressCallback sets the progress callback function
func (e *Executor) SetProgressCallback(callback ProgressCallback) {
	e.progressCallback = callback
}

// SetFailFast sets the fail-fast mode
func (e *Executor) SetFailFast(failFast bool) {
	e.failFast = failFast
}

// NewExecutor creates a new workflow executor
func NewExecutor(cfg *config.Config, log *logger.Logger) *Executor {
	executor := &Executor{
		config:     cfg,
		logger:     log,
		httpClient: httpclient.NewClient(cfg.Timeout, log),
		grpcClient: nil, // Will be initialized when needed
		validator:  validation.NewValidator(log),
		varManager: variables.NewManager(log),
	}

	// Set the variable manager in the validator
	executor.validator.SetVariableManager(executor.varManager)

	return executor
}

// Load loads a workflow from a file
func Load(filename string) (*Workflow, error) {
	return LoadWithImports(filename, []string{"./components", "./templates"})
}

// LoadWithImports loads a workflow from a file with import resolution
func LoadWithImports(filename string, searchPaths []string) (*Workflow, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open workflow file: %w", err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read workflow file: %w", err)
	}

	var workflow Workflow

	// Try YAML first
	if err := yaml.Unmarshal(content, &workflow); err != nil {
		// Try JSON if YAML fails
		if err := json.Unmarshal(content, &workflow); err != nil {
			return nil, fmt.Errorf("failed to parse workflow file: %w", err)
		}
	}

	// Set defaults
	if workflow.Variables == nil {
		workflow.Variables = make(map[string]interface{})
	}

	// Сохраняем путь к исходному workflow-файлу
	workflow.SourceFile = filename

	// Resolve imports if any
	if len(workflow.Imports) > 0 {
		// Get the directory of the workflow file for relative imports
		workflowDir := filepath.Dir(filename)
		searchPaths = append([]string{workflowDir}, searchPaths...)

		componentManager := NewComponentManager(searchPaths)
		if err := componentManager.resolveWorkflowImports(&workflow); err != nil {
			return nil, fmt.Errorf("failed to resolve imports: %w", err)
		}
	}

	return &workflow, nil
}

// Execute executes a workflow and returns the results
func (e *Executor) Execute(wf *Workflow) ([]TestResult, error) {
	e.logger.Info("Starting workflow execution", "name", wf.Name)
	startTime := time.Now()

	// Initialize variables
	e.initializeVariables(wf.Variables)

	// Собираем карту компонент по имени (только step-компоненты)
	componentMap := make(map[string]StepWithVars)
	// Получаем директорию workflow-файла для корректного поиска компонентов
	workflowDir := ""
	if wf != nil && wf.SourceFile != "" {
		workflowDir = filepath.Dir(wf.SourceFile)
	}
	searchPaths := []string{}
	if workflowDir != "" {
		searchPaths = append(searchPaths, workflowDir)
	}
	componentManager := NewComponentManager(searchPaths)
	for _, imp := range wf.Imports {
		component, err := componentManager.LoadComponent(imp.Path)
		if err != nil {
			continue
		}
		if component.Type == "step" && len(component.Steps) == 1 {
			vars := make(map[string]interface{})
			for k, v := range component.Variables {
				vars[k] = v
			}
			for k, v := range imp.Variables {
				vars[k] = v
			}
			componentName := imp.Alias
			if componentName == "" {
				componentName = component.Name
			}
			componentMap[componentName] = StepWithVars{
				Step:      component.Steps[0],
				Variables: vars,
			}
		}
	}

	// DEBUG: выводим все ключи componentMap перед выполнением шагов
	componentKeys := make([]string, 0, len(componentMap))
	for k := range componentMap {
		componentKeys = append(componentKeys, k)
	}
	e.logger.Info("[DEBUG] Available componentMap keys", "keys", componentKeys)

	var allResults []TestResult
	totalSteps := len(wf.Steps)

	for stepIndex, step := range wf.Steps {
		if step.Use != "" {
			if comp, ok := componentMap[step.Use]; ok {
				mergedStep := comp.Step
				if step.Capture != nil {
					mergedStep.Capture = step.Capture
				}
				if len(step.Validate) > 0 {
					mergedStep.Validate = step.Validate
				}
				if step.Name != "" {
					mergedStep.Name = step.Name
				}
				if step.Description != "" {
					mergedStep.Description = step.Description
				}
				if step.Repeat != nil {
					mergedStep.Repeat = step.Repeat
				}
				// Copy Poll configuration: step-level poll overrides component poll
				// If component has poll but step doesn't, keep component's poll
				// If step has poll, it overrides component's poll
				if step.Poll != nil {
					mergedStep.Poll = step.Poll
					e.logger.Debug("[COMPONENT] Overriding poll configuration from step", "step", step.Name, "max_attempts", step.Poll.MaxAttempts, "interval", step.Poll.Interval)
				} else if mergedStep.Poll != nil {
					e.logger.Debug("[COMPONENT] Using poll configuration from component", "step", step.Name, "max_attempts", mergedStep.Poll.MaxAttempts, "interval", mergedStep.Poll.Interval)
				}
				// Note: If step.Poll is nil but component has poll, mergedStep already has it from comp.Step
				// Copy ShowResponse: if either component or step has it set to true, show response
				if step.ShowResponse {
					mergedStep.ShowResponse = true
				}
				// Copy timeout: step-level timeout overrides component timeout
				// If step has timeout in request, it overrides component's request timeout
				if step.Request.Timeout != "" {
					mergedStep.Request.Timeout = step.Request.Timeout
					e.logger.Debug("[COMPONENT] Overriding timeout from step", "step", step.Name, "timeout", step.Request.Timeout)
				} else if step.Timeout != "" {
					// Also check step-level timeout field (for backward compatibility)
					mergedStep.Request.Timeout = step.Timeout
					e.logger.Debug("[COMPONENT] Overriding timeout from step-level field", "step", step.Name, "timeout", step.Timeout)
				} else if mergedStep.Request.Timeout != "" {
					e.logger.Debug("[COMPONENT] Using timeout from component", "step", step.Name, "timeout", mergedStep.Request.Timeout)
				}
				e.logger.Info("[COMPONENT] Executing use step", "use", step.Use, "step", mergedStep.Name)
				// Initialize component variables first
				e.initializeVariables(comp.Variables)
				// Then apply step-level variable overrides
				if len(step.Variables) > 0 {
					e.logger.Debug("[COMPONENT] Applying step-level variable overrides", "variables", step.Variables)
					e.initializeVariables(step.Variables)
				}
				result := &TestResult{
					Name:         mergedStep.Name,
					Status:       "pending",
					CapturedData: make(map[string]interface{}),
				}

				// Notify progress callback that step is starting
				if e.progressCallback != nil {
					e.progressCallback(mergedStep.Name, stepIndex+1, totalSteps, "running", 0, 0, 0, nil)
				}

				if err := e.executeStepWithRepeat(&mergedStep, result); err != nil {
					result.Status = "failed"
					result.Error = err.Error()
					e.logger.Error("Component step execution failed", "component", mergedStep.Name, "error", err)
				} else {
					result.Status = "passed"
				}

				// Notify progress callback of completion
				if e.progressCallback != nil {
					validationsPassed := 0
					validationsTotal := len(result.Validations)
					for _, v := range result.Validations {
						if v.Passed {
							validationsPassed++
						}
					}
					e.progressCallback(mergedStep.Name, stepIndex+1, totalSteps, result.Status, result.Duration, validationsPassed, validationsTotal, nil)
				}

				allResults = append(allResults, *result)
				vars := e.varManager.GetAll()
				e.logger.Info("[DEBUG] Variables after component step", "step", mergedStep.Name, "vars", vars)

				// Check fail-fast mode for component steps
				if e.failFast && result.Status == "failed" {
					e.logger.Error("Fail-fast mode enabled: stopping execution due to component test failure", "step", mergedStep.Name)
					return allResults, fmt.Errorf("execution stopped at first failure (component step: %s)", mergedStep.Name)
				}
				continue
			} else {
				e.logger.Error("Component not found for use step", "use", step.Use)
				result := &TestResult{
					Name:   step.Name,
					Status: "failed",
					Error:  fmt.Sprintf("Component not found for use: %s", step.Use),
				}
				allResults = append(allResults, *result)

				// Check fail-fast mode for missing component
				if e.failFast {
					e.logger.Error("Fail-fast mode enabled: stopping execution due to missing component", "use", step.Use)
					return allResults, fmt.Errorf("execution stopped at first failure (component not found: %s)", step.Use)
				}
				continue
			}
		}
		result := &TestResult{
			Name:         step.Name,
			Status:       "pending",
			CapturedData: make(map[string]interface{}),
		}
		if step.Condition != "" {
			if !e.evaluateCondition(step.Condition) {
				e.logger.Info("Skipping step due to condition", "step", step.Name, "condition", step.Condition)
				result.Status = "skipped"
				allResults = append(allResults, *result)
				continue
			}
		}

		// Notify progress callback that step is starting
		if e.progressCallback != nil {
			e.progressCallback(step.Name, stepIndex+1, totalSteps, "running", 0, 0, 0, nil)
		}

		if err := e.executeStepWithRepeat(&step, result); err != nil {
			result.Status = "failed"
			result.Error = err.Error()
			e.logger.Error("Step execution failed", "step", step.Name, "error", err)
		} else {
			result.Status = "passed"
		}

		// Notify progress callback of completion
		if e.progressCallback != nil {
			validationsPassed := 0
			validationsTotal := len(result.Validations)
			for _, v := range result.Validations {
				if v.Passed {
					validationsPassed++
				}
			}
			var resultErr error
			if result.Status == "failed" && result.Error != "" {
				resultErr = fmt.Errorf("%s", result.Error)
			}
			e.progressCallback(step.Name, stepIndex+1, totalSteps, result.Status, result.Duration, validationsPassed, validationsTotal, resultErr)
		}

		allResults = append(allResults, *result)
		vars := e.varManager.GetAll()
		e.logger.Info("[DEBUG] Variables after step", "step", step.Name, "vars", vars)

		// Check fail-fast mode
		if e.failFast && result.Status == "failed" {
			e.logger.Error("Fail-fast mode enabled: stopping execution due to test failure", "step", step.Name)
			return allResults, fmt.Errorf("execution stopped at first failure (step: %s)", step.Name)
		}
	}

	// Execute step groups
	for _, group := range wf.Groups {
		groupResult := &GroupResult{
			Name:     group.Name,
			Status:   "pending",
			Parallel: group.Parallel,
		}

		// Check condition if specified
		if group.Condition != "" {
			if !e.evaluateCondition(group.Condition) {
				e.logger.Info("Skipping group due to condition", "group", group.Name, "condition", group.Condition)
				groupResult.Status = "skipped"
				// Convert group result to test results
				for _, step := range group.Steps {
					allResults = append(allResults, TestResult{
						Name:   fmt.Sprintf("%s.%s", group.Name, step.Name),
						Status: "skipped",
					})
				}
				continue
			}
		}

		groupStartTime := time.Now()
		if err := e.executeGroup(&group, groupResult); err != nil {
			groupResult.Status = "failed"
			groupResult.Error = err.Error()
			e.logger.Error("Group execution failed", "group", group.Name, "error", err)
		} else {
			groupResult.Status = "passed"
		}
		groupResult.Duration = time.Since(groupStartTime)

		// Add group results to all results
		for _, result := range groupResult.Results {
			result.Name = fmt.Sprintf("%s.%s", group.Name, result.Name)
			allResults = append(allResults, result)
		}
	}

	totalDuration := time.Since(startTime)
	e.logger.Info("Workflow execution completed", "duration", totalDuration, "total_steps", len(allResults))

	return allResults, nil
}

// executeGroup executes a group of steps, either sequentially or in parallel
func (e *Executor) executeGroup(group *StepGroup, groupResult *GroupResult) error {
	e.logger.Info("Executing step group", "group", group.Name, "parallel", group.Parallel, "steps", len(group.Steps))

	if group.Parallel {
		return e.executeGroupParallel(group, groupResult)
	}
	return e.executeGroupSequential(group, groupResult)
}

// executeGroupSequential executes steps in a group sequentially
func (e *Executor) executeGroupSequential(group *StepGroup, groupResult *GroupResult) error {
	for _, step := range group.Steps {
		result := &TestResult{
			Name:         step.Name,
			Status:       "pending",
			CapturedData: make(map[string]interface{}),
		}

		// Check condition if specified
		if step.Condition != "" {
			if !e.evaluateCondition(step.Condition) {
				e.logger.Info("Skipping step due to condition", "step", step.Name, "condition", step.Condition)
				result.Status = "skipped"
				groupResult.Results = append(groupResult.Results, *result)
				continue
			}
		}

		if err := e.executeStepWithRepeat(&step, result); err != nil {
			result.Status = "failed"
			result.Error = err.Error()
			e.logger.Error("Step execution failed", "step", step.Name, "error", err)
		} else {
			result.Status = "passed"
		}

		groupResult.Results = append(groupResult.Results, *result)
	}

	return nil
}

// executeGroupParallel executes steps in a group in parallel
func (e *Executor) executeGroupParallel(group *StepGroup, groupResult *GroupResult) error {
	var wg sync.WaitGroup
	results := make([]*TestResult, len(group.Steps))
	errors := make(chan error, len(group.Steps))

	for i, step := range group.Steps {
		wg.Add(1)
		go func(stepIndex int, step Step) {
			defer wg.Done()

			result := &TestResult{
				Name:         step.Name,
				Status:       "pending",
				CapturedData: make(map[string]interface{}),
			}

			// Check condition if specified
			if step.Condition != "" {
				if !e.evaluateCondition(step.Condition) {
					e.logger.Info("Skipping step due to condition", "step", step.Name, "condition", step.Condition)
					result.Status = "skipped"
					results[stepIndex] = result
					return
				}
			}

			if err := e.executeStepWithRepeat(&step, result); err != nil {
				result.Status = "failed"
				result.Error = err.Error()
				e.logger.Error("Step execution failed", "step", step.Name, "error", err)
				errors <- err
			} else {
				result.Status = "passed"
			}

			results[stepIndex] = result
		}(i, step)
	}

	wg.Wait()
	close(errors)

	// Check for any errors
	select {
	case err := <-errors:
		return err
	default:
		// No errors
	}

	// Add results to group result
	for _, result := range results {
		if result != nil {
			groupResult.Results = append(groupResult.Results, *result)
		}
	}

	return nil
}

// evaluateCondition evaluates a condition expression
func (e *Executor) evaluateCondition(condition string) bool {
	// Simple condition evaluation - can be extended for more complex logic
	// For now, we'll support basic variable checks
	if strings.HasPrefix(condition, "{{") && strings.HasSuffix(condition, "}}") {
		// Extract variable name
		varName := strings.TrimSpace(condition[2 : len(condition)-2])
		value, exists := e.varManager.Get(varName)

		// Check if variable exists and has a truthy value
		if exists && value != nil {
			switch v := value.(type) {
			case bool:
				return v
			case string:
				return v != "" && v != "false" && v != "0"
			case int, int32, int64, float32, float64:
				return v != 0
			default:
				return value != nil
			}
		}
		return false
	}

	// For now, assume condition is true if it's not a variable reference
	return true
}

// executeStep executes a single step with retry logic
func (e *Executor) executeStep(step *Step, result *TestResult) error {
	if result.CapturedData == nil {
		result.CapturedData = make(map[string]interface{})
	}
	startTime := time.Now()
	maxRetries := step.Retry
	if maxRetries == 0 {
		maxRetries = 1 // Default to 1 attempt
	}

	// Всегда сохраняем print-текст, если он есть
	if step.Print != "" {
		msg, _ := e.varManager.Substitute(step.Print)
		result.PrintText = msg
	}

	// Print-only step (нет запроса, wait, use)
	if step.Print != "" && step.Request.Method == "" && step.Request.URL == "" && step.Request.Service == "" && step.Wait == "" && step.Use == "" {
		result.Status = "passed"
		result.Duration = time.Since(startTime)
		return nil
	}

	// Check if this is a wait-only step (no request)
	if step.Wait != "" && step.Request.Method == "" && step.Request.URL == "" && step.Request.Service == "" {
		// This is a wait-only step, just wait and return success
		duration := e.parseTimeout(step.Wait)
		if duration > 0 {
			e.logger.Info("Executing wait step", "step", step.Name, "wait", duration)
			time.Sleep(duration)
		}
		result.Duration = time.Since(startTime)
		return nil
	}

	// Check if this step has polling configuration
	if step.Poll != nil {
		return e.executeStepWithPoll(step, result, startTime)
	}

	var lastError error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			result.Retries = attempt
			retryDelay := e.parseTimeout(step.RetryDelay)
			if retryDelay > 0 {
				e.logger.Info("Retrying step", "step", step.Name, "attempt", attempt+1, "delay", retryDelay)
				time.Sleep(retryDelay)
			}
		}

		// Substitute variables in request
		substitutedReq, err := e.substituteRequestVariables(&step.Request)
		if err != nil {
			lastError = fmt.Errorf("variable substitution failed: %w", err)
			continue
		}

		// Set default protocol to HTTP if not specified
		if substitutedReq.Protocol == "" {
			substitutedReq.Protocol = "http"
		}

		e.logger.Debug("Request details",
			"protocol", substitutedReq.Protocol,
			"url", substitutedReq.URL,
			"service", substitutedReq.Service,
			"method", substitutedReq.Method,
			"grpc_method", substitutedReq.GRPCMethod)

		// Execute request based on protocol
		var httpResponse *httpclient.Response
		var grpcResponse *grpcclient.Response
		var dbResponse *dbclient.Response
		var requestErr error

		e.logger.Debug("Executing request", "protocol", substitutedReq.Protocol)

		if substitutedReq.Protocol == "grpc" {
			// Initialize gRPC client if not already done
			if e.grpcClient == nil {
				grpcClient, err := grpcclient.NewClient(substitutedReq.ServerAddr, substitutedReq.Insecure, e.logger)
				if err != nil {
					lastError = fmt.Errorf("failed to create gRPC client: %w", err)
					continue
				}
				e.grpcClient = grpcClient
			}

			// Execute gRPC request
			grpcReq := &grpcclient.Request{
				Service:    substitutedReq.Service,
				Method:     substitutedReq.GRPCMethod,
				Data:       substitutedReq.Data,
				Metadata:   substitutedReq.Metadata,
				ServerAddr: substitutedReq.ServerAddr,
				Insecure:   substitutedReq.Insecure,
				Timeout:    e.parseTimeout(substitutedReq.Timeout),
			}
			grpcResponse, requestErr = e.grpcClient.Execute(grpcReq)
		} else if substitutedReq.Protocol == "db" {
			// Execute database request
			if substitutedReq.DBConfig == nil {
				lastError = fmt.Errorf("database configuration is required for db protocol")
				continue
			}

			// Substitute variables in database config
			dbConfig := substitutedReq.DBConfig
			// If DSN is provided, substitute variables in DSN
			if dbConfig.DSN != "" {
				if subDSN, err := e.varManager.Substitute(dbConfig.DSN); err == nil {
					dbConfig.DSN = subDSN
				}
			} else {
				// Otherwise, substitute individual parameters
				if dbConfig.Host != "" {
					if subHost, err := e.varManager.Substitute(dbConfig.Host); err == nil {
						dbConfig.Host = subHost
					}
				}
				if dbConfig.Username != "" {
					if subUser, err := e.varManager.Substitute(dbConfig.Username); err == nil {
						dbConfig.Username = subUser
					}
				}
				if dbConfig.Password != "" {
					if subPass, err := e.varManager.Substitute(dbConfig.Password); err == nil {
						dbConfig.Password = subPass
					}
				}
				if dbConfig.Database != "" {
					if subDB, err := e.varManager.Substitute(dbConfig.Database); err == nil {
						dbConfig.Database = subDB
					}
				}
			}

			// Get query string
			query := ""
			if queryStr, ok := substitutedReq.Query.(string); ok {
				query = queryStr
			}
			if query == "" {
				lastError = fmt.Errorf("query is required for db protocol")
				continue
			}

			// Set timeout if not set
			if dbConfig.Timeout == 0 {
				dbConfig.Timeout = e.parseTimeout(substitutedReq.Timeout)
			}

			// Create or reuse database client
			// For now, create a new client each time (could be optimized to reuse connections)
			dbClient, err := dbclient.NewClient(dbConfig, e.logger)
			if err != nil {
				lastError = fmt.Errorf("failed to create database client: %w", err)
				continue
			}
			defer dbClient.Close()

			dbResponse, requestErr = dbClient.Execute(query)
		} else {
			// Execute HTTP request (default)
			queryMap := make(map[string]string)
			if query, ok := substitutedReq.Query.(map[string]string); ok {
				queryMap = query
			}
			httpReq := &httpclient.Request{
				Method:  substitutedReq.Method,
				URL:     substitutedReq.URL,
				Headers: substitutedReq.Headers,
				Body:    substitutedReq.Body,
				Query:   queryMap,
				Timeout: e.parseTimeout(substitutedReq.Timeout),
				Auth:    substitutedReq.Auth,
			}
			httpResponse, requestErr = e.httpClient.Execute(httpReq)
		}

		// Show response if requested (always show, even on errors)
		if step.ShowResponse {
			if substitutedReq.Protocol == "grpc" {
				if grpcResponse != nil {
					jsonData, err := json.MarshalIndent(grpcResponse.Data, "", "  ")
					if err == nil {
						fmt.Println("================ RESPONSE (gRPC) ================")
						fmt.Println(string(jsonData))
						fmt.Println("================ END RESPONSE ================")
					}
				}
			} else if substitutedReq.Protocol == "db" {
				if dbResponse != nil {
					jsonData, err := json.MarshalIndent(dbResponse.Data, "", "  ")
					if err == nil {
						fmt.Println("================ RESPONSE (DB) ================")
						fmt.Println(string(jsonData))
						fmt.Println("================ END RESPONSE ================")
					}
				}
			} else {
				if httpResponse != nil && len(httpResponse.Body) > 0 {
					fmt.Println("================ RESPONSE ================")
					fmt.Println(string(httpResponse.Body))
					fmt.Println("================ END RESPONSE ================")
				}
			}
		}

		if requestErr != nil {
			lastError = fmt.Errorf("request failed: %w", requestErr)
			// Log API response on request error when verbose is enabled (if response exists)
			if !e.logger.IsMuted() {
				e.logAPIResponseOnFailure(substitutedReq.Protocol, httpResponse, grpcResponse, dbResponse, step.Name)
			}
			continue
		}

		// Run validations
		var validationErrors []string
		if len(step.Validate) > 0 {
			var validationResults []validation.ValidationResult
			var validationErr error

			if substitutedReq.Protocol == "grpc" {
				// For gRPC, we'll need to create a mock HTTP response for validation
				// This is a simplified approach - in production you'd want proper gRPC validation
				jsonData, err := json.Marshal(grpcResponse.Data)
				if err != nil {
					validationErr = fmt.Errorf("failed to marshal gRPC response for validation: %w", err)
				} else {
					mockResponse := &httpclient.Response{
						StatusCode: 200, // gRPC OK status
						Body:       jsonData,
						Duration:   grpcResponse.Duration,
					}
					validationResults, validationErr = e.validator.Validate(mockResponse, step.Validate)
				}
			} else if substitutedReq.Protocol == "db" {
				// For database, create a mock HTTP response for validation
				jsonData, err := json.Marshal(dbResponse.Data)
				if err != nil {
					validationErr = fmt.Errorf("failed to marshal database response for validation: %w", err)
				} else {
					mockResponse := &httpclient.Response{
						StatusCode: 200, // Database OK status
						Body:       jsonData,
						Duration:   dbResponse.Duration,
					}
					validationResults, validationErr = e.validator.Validate(mockResponse, step.Validate)
				}
			} else {
				validationResults, validationErr = e.validator.Validate(httpResponse, step.Validate)
			}

			// Always save validation results for CLI output
			result.Validations = validationResults

			if validationErr != nil {
				validationErrors = append(validationErrors, validationErr.Error())
			} else {
				for _, validationResult := range validationResults {
					if !validationResult.Passed {
						validationErrors = append(validationErrors, validationResult.Error)
					}
				}
			}
		}

		if len(validationErrors) > 0 {
			lastError = fmt.Errorf("validation failed: %s", strings.Join(validationErrors, "; "))
			// Log API response on validation failure when verbose is enabled
			if !e.logger.IsMuted() {
				e.logAPIResponseOnFailure(substitutedReq.Protocol, httpResponse, grpcResponse, dbResponse, step.Name)
			}
			continue
		}

		// Capture values if specified
		if step.Capture != nil {
			var captureErr error
			if substitutedReq.Protocol == "grpc" {
				// For gRPC, create a mock response for capture
				jsonData, err := json.Marshal(grpcResponse.Data)
				if err != nil {
					captureErr = fmt.Errorf("failed to marshal gRPC response: %w", err)
				} else {
					mockResponse := &httpclient.Response{
						StatusCode: 200,
						Body:       jsonData,
						Duration:   grpcResponse.Duration,
					}
					captureErr = e.captureValues(mockResponse, step.Capture, result)
				}
			} else if substitutedReq.Protocol == "db" {
				// For database, create a mock response for capture
				jsonData, err := json.Marshal(dbResponse.Data)
				if err != nil {
					captureErr = fmt.Errorf("failed to marshal database response: %w", err)
				} else {
					mockResponse := &httpclient.Response{
						StatusCode: 200,
						Body:       jsonData,
						Duration:   dbResponse.Duration,
					}
					captureErr = e.captureValues(mockResponse, step.Capture, result)
				}
			} else {
				captureErr = e.captureValues(httpResponse, step.Capture, result)
			}
			if captureErr != nil {
				e.logger.Warn("Failed to capture values", "step", step.Name, "error", captureErr)
			}
		}

		// Success - no need to retry
		result.Duration = time.Since(startTime)
		return nil
	}

	result.Duration = time.Since(startTime)
	return lastError
}

// executeStepWithPoll executes a step with polling until conditions are met
func (e *Executor) executeStepWithPoll(step *Step, result *TestResult, startTime time.Time) error {
	pollConfig := step.Poll
	maxAttempts := pollConfig.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 10 // Default to 10 attempts if not specified
	}

	interval := e.parseTimeout(pollConfig.Interval)
	if interval <= 0 {
		interval = 1 * time.Second // Default to 1 second if not specified
	}

	e.logger.Info("Executing step with polling",
		"step", step.Name,
		"max_attempts", maxAttempts,
		"interval", interval)

	var lastError error
	var lastHTTPResponse *httpclient.Response
	var lastGRPCResponse *grpcclient.Response
	var lastDBResponse *dbclient.Response
	var lastSubstitutedReq *Request

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		e.logger.Debug("Polling attempt", "step", step.Name, "attempt", attempt, "max_attempts", maxAttempts)

		// Substitute variables in request
		substitutedReq, err := e.substituteRequestVariables(&step.Request)
		if err == nil {
			lastSubstitutedReq = substitutedReq
		}
		if err != nil {
			lastError = fmt.Errorf("variable substitution failed: %w", err)
			if attempt < maxAttempts {
				time.Sleep(interval)
			}
			continue
		}

		// Set default protocol to HTTP if not specified
		if substitutedReq.Protocol == "" {
			substitutedReq.Protocol = "http"
		}

		// Execute request based on protocol
		var httpResponse *httpclient.Response
		var grpcResponse *grpcclient.Response
		var dbResponse *dbclient.Response
		var requestErr error

		if substitutedReq.Protocol == "grpc" {
			// Initialize gRPC client if not already done
			if e.grpcClient == nil {
				grpcClient, err := grpcclient.NewClient(substitutedReq.ServerAddr, substitutedReq.Insecure, e.logger)
				if err != nil {
					lastError = fmt.Errorf("failed to create gRPC client: %w", err)
					if attempt < maxAttempts {
						time.Sleep(interval)
					}
					continue
				}
				e.grpcClient = grpcClient
			}

			// Execute gRPC request
			grpcReq := &grpcclient.Request{
				Service:    substitutedReq.Service,
				Method:     substitutedReq.GRPCMethod,
				Data:       substitutedReq.Data,
				Metadata:   substitutedReq.Metadata,
				ServerAddr: substitutedReq.ServerAddr,
				Insecure:   substitutedReq.Insecure,
				Timeout:    e.parseTimeout(substitutedReq.Timeout),
			}
			grpcResponse, requestErr = e.grpcClient.Execute(grpcReq)
			if grpcResponse != nil {
				lastGRPCResponse = grpcResponse
			}
		} else if substitutedReq.Protocol == "db" {
			// Execute database request
			if substitutedReq.DBConfig == nil {
				lastError = fmt.Errorf("database configuration is required for db protocol")
				if attempt < maxAttempts {
					time.Sleep(interval)
				}
				continue
			}

			// Substitute variables in database config
			dbConfig := substitutedReq.DBConfig
			// If DSN is provided, substitute variables in DSN
			if dbConfig.DSN != "" {
				if subDSN, err := e.varManager.Substitute(dbConfig.DSN); err == nil {
					dbConfig.DSN = subDSN
				}
			} else {
				// Otherwise, substitute individual parameters
				if dbConfig.Host != "" {
					if subHost, err := e.varManager.Substitute(dbConfig.Host); err == nil {
						dbConfig.Host = subHost
					}
				}
				if dbConfig.Username != "" {
					if subUser, err := e.varManager.Substitute(dbConfig.Username); err == nil {
						dbConfig.Username = subUser
					}
				}
				if dbConfig.Password != "" {
					if subPass, err := e.varManager.Substitute(dbConfig.Password); err == nil {
						dbConfig.Password = subPass
					}
				}
				if dbConfig.Database != "" {
					if subDB, err := e.varManager.Substitute(dbConfig.Database); err == nil {
						dbConfig.Database = subDB
					}
				}
			}

			// Get query string
			query := ""
			if queryStr, ok := substitutedReq.Query.(string); ok {
				query = queryStr
			}
			if query == "" {
				lastError = fmt.Errorf("query is required for db protocol")
				if attempt < maxAttempts {
					time.Sleep(interval)
				}
				continue
			}

			// Set timeout if not set
			if dbConfig.Timeout == 0 {
				dbConfig.Timeout = e.parseTimeout(substitutedReq.Timeout)
			}

			// Create database client
			dbClient, err := dbclient.NewClient(dbConfig, e.logger)
			if err != nil {
				lastError = fmt.Errorf("failed to create database client: %w", err)
				if attempt < maxAttempts {
					time.Sleep(interval)
				}
				continue
			}
			defer dbClient.Close()

			dbResponse, requestErr = dbClient.Execute(query)
			if dbResponse != nil {
				lastDBResponse = dbResponse
			}
		} else {
			// Execute HTTP request (default)
			queryMap := make(map[string]string)
			if query, ok := substitutedReq.Query.(map[string]string); ok {
				queryMap = query
			}
			httpReq := &httpclient.Request{
				Method:  substitutedReq.Method,
				URL:     substitutedReq.URL,
				Headers: substitutedReq.Headers,
				Body:    substitutedReq.Body,
				Query:   queryMap,
				Timeout: e.parseTimeout(substitutedReq.Timeout),
				Auth:    substitutedReq.Auth,
			}
			httpResponse, requestErr = e.httpClient.Execute(httpReq)
			if httpResponse != nil {
				lastHTTPResponse = httpResponse
			}
		}

		// Show response if requested (always show, even on errors)
		if step.ShowResponse {
			if substitutedReq.Protocol == "grpc" {
				if grpcResponse != nil {
					jsonData, err := json.MarshalIndent(grpcResponse.Data, "", "  ")
					if err == nil {
						fmt.Println("================ RESPONSE (gRPC) ================")
						fmt.Println(string(jsonData))
						fmt.Println("================ END RESPONSE ================")
					}
				}
			} else if substitutedReq.Protocol == "db" {
				if dbResponse != nil {
					jsonData, err := json.MarshalIndent(dbResponse.Data, "", "  ")
					if err == nil {
						fmt.Println("================ RESPONSE (DB) ================")
						fmt.Println(string(jsonData))
						fmt.Println("================ END RESPONSE ================")
					}
				}
			} else {
				if httpResponse != nil && len(httpResponse.Body) > 0 {
					fmt.Println("================ RESPONSE ================")
					fmt.Println(string(httpResponse.Body))
					fmt.Println("================ END RESPONSE ================")
				}
			}
		}

		if requestErr != nil {
			lastError = fmt.Errorf("request failed: %w", requestErr)
			e.logger.Debug("Polling attempt failed", "step", step.Name, "attempt", attempt, "error", requestErr)
			// Log API response on request error when verbose is enabled (if response exists)
			if !e.logger.IsMuted() {
				e.logAPIResponseOnFailure(substitutedReq.Protocol, httpResponse, grpcResponse, dbResponse, step.Name)
			}
			if attempt < maxAttempts {
				time.Sleep(interval)
			}
			continue
		}

		// Check polling conditions (poll.until)
		var validationResults []validation.ValidationResult
		var validationErr error
		var responseForValidation *httpclient.Response

		if substitutedReq.Protocol == "grpc" {
			// For gRPC, create a mock HTTP response for validation
			jsonData, err := json.Marshal(grpcResponse.Data)
			if err != nil {
				lastError = fmt.Errorf("failed to marshal gRPC response for validation: %w", err)
				if attempt < maxAttempts {
					time.Sleep(interval)
				}
				continue
			}
			responseForValidation = &httpclient.Response{
				StatusCode: 200, // gRPC OK status
				Body:       jsonData,
				Duration:   grpcResponse.Duration,
			}
		} else if substitutedReq.Protocol == "db" {
			// For database, create a mock HTTP response for validation
			jsonData, err := json.Marshal(dbResponse.Data)
			if err != nil {
				lastError = fmt.Errorf("failed to marshal database response for validation: %w", err)
				if attempt < maxAttempts {
					time.Sleep(interval)
				}
				continue
			}
			responseForValidation = &httpclient.Response{
				StatusCode: 200, // Database OK status
				Body:       jsonData,
				Duration:   dbResponse.Duration,
			}
		} else {
			responseForValidation = httpResponse
		}

		// Validate against poll.until conditions
		validationResults, validationErr = e.validator.Validate(responseForValidation, pollConfig.Until)

		// Also run regular validations if specified (for reporting)
		if len(step.Validate) > 0 {
			regularResults, _ := e.validator.Validate(responseForValidation, step.Validate)
			result.Validations = regularResults
		} else {
			// Use polling validation results for reporting
			result.Validations = validationResults
		}

		// Check if all polling conditions are met
		allConditionsMet := true
		if validationErr != nil {
			allConditionsMet = false
		} else {
			for _, validationResult := range validationResults {
				if !validationResult.Passed {
					allConditionsMet = false
					break
				}
			}
		}

		if allConditionsMet {
			// Success! All polling conditions are met
			result.PollAttempts = attempt
			e.logger.Info("Polling condition met", "step", step.Name, "attempt", attempt, "total_attempts", attempt)

			// Capture values if specified
			if step.Capture != nil {
				var captureErr error
				if substitutedReq.Protocol == "grpc" || substitutedReq.Protocol == "db" {
					captureErr = e.captureValues(responseForValidation, step.Capture, result)
				} else {
					captureErr = e.captureValues(httpResponse, step.Capture, result)
				}
				if captureErr != nil {
					e.logger.Warn("Failed to capture values", "step", step.Name, "error", captureErr)
				}
			}

			result.Duration = time.Since(startTime)
			return nil
		}

		// Conditions not met, log and continue polling
		e.logger.Debug("Polling condition not met, will retry",
			"step", step.Name,
			"attempt", attempt,
			"max_attempts", maxAttempts)

		// If this is not the last attempt, wait before next poll
		if attempt < maxAttempts {
			time.Sleep(interval)
		}
	}

	// All attempts exhausted, condition not met
	result.PollAttempts = maxAttempts
	result.Duration = time.Since(startTime)

	// Log API response on failure when verbose is enabled
	if !e.logger.IsMuted() && lastSubstitutedReq != nil {
		e.logAPIResponseOnFailure(lastSubstitutedReq.Protocol, lastHTTPResponse, lastGRPCResponse, lastDBResponse, step.Name)
	}

	if lastError != nil {
		return fmt.Errorf("polling failed after %d attempts: %w", maxAttempts, lastError)
	}

	// Build error message from failed validations
	var failedValidations []string
	for _, validationResult := range result.Validations {
		if !validationResult.Passed {
			failedValidations = append(failedValidations, validationResult.Error)
		}
	}

	if len(failedValidations) > 0 {
		return fmt.Errorf("polling condition not met after %d attempts: %s", maxAttempts, strings.Join(failedValidations, "; "))
	}

	return fmt.Errorf("polling condition not met after %d attempts", maxAttempts)
}

// executeStepWithRepeat executes a step with repeat configuration
func (e *Executor) executeStepWithRepeat(step *Step, result *TestResult) error {
	if step.Repeat == nil {
		// No repeat configuration, execute normally
		return e.executeStep(step, result)
	}

	repeatConfig := step.Repeat
	result.RepeatCount = repeatConfig.Count
	result.RepeatResults = make([]TestResult, 0, repeatConfig.Count)

	e.logger.Info("Executing step with repeat",
		"step", step.Name,
		"count", repeatConfig.Count,
		"parallel", repeatConfig.Parallel)

	if repeatConfig.Parallel {
		return e.executeStepRepeatParallel(step, result, repeatConfig)
	} else {
		return e.executeStepRepeatSequential(step, result, repeatConfig)
	}
}

// executeStepRepeatSequential executes a step multiple times sequentially
func (e *Executor) executeStepRepeatSequential(step *Step, result *TestResult, repeatConfig *RepeatConfig) error {
	delay := e.parseTimeout(repeatConfig.Delay)

	for i := 0; i < repeatConfig.Count; i++ {
		e.logger.Debug("Executing repeat iteration",
			"step", step.Name,
			"iteration", i+1,
			"total", repeatConfig.Count)

		// Create a copy of the step for this iteration
		stepCopy := *step

		// Apply repeat variables if specified
		if repeatConfig.Variables != nil || true {
			// Create a temporary variable manager for this iteration
			tempVarManager := variables.NewManager(e.logger)

			// Copy current variables
			for key, value := range e.varManager.GetAll() {
				tempVarManager.Set(key, value)
			}

			// Автоматически добавляем iteration и index
			tempVarManager.Set("iteration", i+1)
			tempVarManager.Set("index", i)

			// Apply repeat-specific variables
			if repeatConfig.Variables != nil {
				for key, value := range repeatConfig.Variables {
					if strValue, ok := value.(string); ok {
						strValue = strings.ReplaceAll(strValue, "{{index}}", strconv.Itoa(i))
						strValue = strings.ReplaceAll(strValue, "{{iteration}}", strconv.Itoa(i+1))
						tempVarManager.Set(key, strValue)
					} else {
						tempVarManager.Set(key, value)
					}
				}
			}

			// Temporarily replace the variable manager
			originalVarManager := e.varManager
			e.varManager = tempVarManager
			defer func() { e.varManager = originalVarManager }()
		}

		// Execute the step
		iterationResult := &TestResult{
			Name:         fmt.Sprintf("%s (iteration %d)", step.Name, i+1),
			CapturedData: make(map[string]interface{}),
		}

		err := e.executeStep(&stepCopy, iterationResult)
		if err != nil {
			iterationResult.Status = "failed"
			iterationResult.Error = err.Error()
		} else {
			iterationResult.Status = "passed"
		}

		result.RepeatResults = append(result.RepeatResults, *iterationResult)

		// Add delay between iterations (except for the last one)
		if i < repeatConfig.Count-1 && delay > 0 {
			e.logger.Debug("Waiting between iterations", "delay", delay)
			time.Sleep(delay)
		}
	}

	// Determine overall result
	passedCount := 0
	for _, repeatResult := range result.RepeatResults {
		if repeatResult.Status == "passed" {
			passedCount++
		}
	}

	if passedCount == repeatConfig.Count {
		result.Status = "passed"
	} else {
		result.Status = "failed"
		result.Error = fmt.Sprintf("%d/%d iterations failed", repeatConfig.Count-passedCount, repeatConfig.Count)
	}

	return nil
}

// executeStepRepeatParallel executes a step multiple times in parallel
func (e *Executor) executeStepRepeatParallel(step *Step, result *TestResult, repeatConfig *RepeatConfig) error {
	results := make([]TestResult, repeatConfig.Count)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for i := 0; i < repeatConfig.Count; i++ {
		wg.Add(1)
		go func(iteration int) {
			defer wg.Done()

			e.logger.Debug("Executing parallel repeat iteration",
				"step", step.Name,
				"iteration", iteration+1,
				"total", repeatConfig.Count)

			// Create a copy of the step for this iteration
			stepCopy := *step

			// Apply repeat variables if specified
			if repeatConfig.Variables != nil {
				// Create a temporary variable manager for this iteration
				tempVarManager := variables.NewManager(e.logger)

				// Copy current variables
				mu.Lock()
				for key, value := range e.varManager.GetAll() {
					tempVarManager.Set(key, value)
				}
				mu.Unlock()

				// Apply repeat-specific variables
				for key, value := range repeatConfig.Variables {
					// Replace {{index}} with current iteration index
					if strValue, ok := value.(string); ok {
						strValue = strings.ReplaceAll(strValue, "{{index}}", strconv.Itoa(iteration))
						strValue = strings.ReplaceAll(strValue, "{{iteration}}", strconv.Itoa(iteration+1))
						tempVarManager.Set(key, strValue)
					} else {
						tempVarManager.Set(key, value)
					}
				}

				// Temporarily replace the variable manager
				mu.Lock()
				originalVarManager := e.varManager
				e.varManager = tempVarManager
				mu.Unlock()
				defer func() {
					mu.Lock()
					e.varManager = originalVarManager
					mu.Unlock()
				}()
			}

			// Execute the step
			iterationResult := &TestResult{
				Name:         fmt.Sprintf("%s (iteration %d)", step.Name, iteration+1),
				CapturedData: make(map[string]interface{}),
			}

			err := e.executeStep(&stepCopy, iterationResult)
			if err != nil {
				iterationResult.Status = "failed"
				iterationResult.Error = err.Error()
			} else {
				iterationResult.Status = "passed"
			}

			mu.Lock()
			results[iteration] = *iterationResult
			mu.Unlock()
		}(i)
	}

	wg.Wait()

	// Collect results
	result.RepeatResults = results

	// Determine overall result
	passedCount := 0
	for _, repeatResult := range result.RepeatResults {
		if repeatResult.Status == "passed" {
			passedCount++
		}
	}

	if passedCount == repeatConfig.Count {
		result.Status = "passed"
	} else {
		result.Status = "failed"
		result.Error = fmt.Sprintf("%d/%d iterations failed", repeatConfig.Count-passedCount, repeatConfig.Count)
	}

	return nil
}

// initializeVariables initializes the variable manager with workflow variables
func (e *Executor) initializeVariables(vars map[string]interface{}) {
	for key, value := range vars {
		e.varManager.Set(key, value)
	}
}

// substituteRequestVariables substitutes variables in the request
func (e *Executor) substituteRequestVariables(req *Request) (*Request, error) {
	e.logger.Debug("Substituting variables in request", "original_url", req.URL)

	substituted := &Request{
		Protocol:   req.Protocol,
		Method:     req.Method,
		URL:        req.URL,
		Headers:    make(map[string]string),
		Body:       req.Body,
		Query:      req.Query, // Can be string for DB or map for HTTP
		Service:    req.Service,
		GRPCMethod: req.GRPCMethod,
		Data:       req.Data,
		Metadata:   make(map[string]string),
		ServerAddr: req.ServerAddr,
		Insecure:   req.Insecure,
		DBConfig:   req.DBConfig,
		Timeout:    req.Timeout,
	}

	// Substitute URL
	if substitutedURL, err := e.varManager.Substitute(req.URL); err != nil {
		e.logger.Error("Failed to substitute URL", "url", req.URL, "error", err)
		return nil, fmt.Errorf("failed to substitute URL: %w", err)
	} else {
		substituted.URL = substitutedURL
		e.logger.Debug("URL substitution result", "original", req.URL, "substituted", substitutedURL)
	}

	// Substitute headers
	for key, value := range req.Headers {
		if substitutedValue, err := e.varManager.Substitute(value); err != nil {
			e.logger.Error("Failed to substitute header", "key", key, "value", value, "error", err)
			return nil, fmt.Errorf("failed to substitute header %s: %w", key, err)
		} else {
			substituted.Headers[key] = substitutedValue
			e.logger.Debug("Header substitution result", "key", key, "original", value, "substituted", substitutedValue)
		}
	}

	// Substitute query parameters (for HTTP) or query string (for DB)
	if req.Protocol == "db" {
		// For database, Query is a string (SQL query)
		if queryStr, ok := req.Query.(string); ok && queryStr != "" {
			if substitutedQuery, err := e.varManager.Substitute(queryStr); err != nil {
				e.logger.Error("Failed to substitute database query", "query", queryStr, "error", err)
				return nil, fmt.Errorf("failed to substitute database query: %w", err)
			} else {
				substituted.Query = substitutedQuery
				e.logger.Debug("Database query substitution result", "original", queryStr, "substituted", substitutedQuery)
			}
		} else {
			substituted.Query = req.Query
		}
	} else {
		// For HTTP, Query is a map
		if queryMap, ok := req.Query.(map[string]string); ok {
			substituted.Query = make(map[string]string)
			for key, value := range queryMap {
				if substitutedValue, err := e.varManager.Substitute(value); err != nil {
					e.logger.Error("Failed to substitute query", "key", key, "value", value, "error", err)
					return nil, fmt.Errorf("failed to substitute query %s: %w", key, err)
				} else {
					substituted.Query.(map[string]string)[key] = substitutedValue
					e.logger.Debug("Query substitution result", "key", key, "original", value, "substituted", substitutedValue)
				}
			}
		} else {
			substituted.Query = req.Query
		}
	}

	// Substitute body
	if req.Body != nil {
		switch body := req.Body.(type) {
		case string:
			if substitutedBody, err := e.varManager.Substitute(body); err != nil {
				e.logger.Error("Failed to substitute body", "body", body, "error", err)
				return nil, fmt.Errorf("failed to substitute body: %w", err)
			} else {
				substituted.Body = substitutedBody
				e.logger.Debug("Body substitution result", "original", body, "substituted", substitutedBody)
			}
		case map[string]interface{}:
			if substitutedBody, err := e.varManager.SubstituteMap(body); err != nil {
				e.logger.Error("Failed to substitute body map", "body", body, "error", err)
				return nil, fmt.Errorf("failed to substitute body: %w", err)
			} else {
				substituted.Body = substitutedBody
				e.logger.Debug("Body map substitution result", "original", body, "substituted", substitutedBody)
			}
		default:
			substituted.Body = req.Body
		}
	}

	// Substitute gRPC fields
	if substitutedServerAddr, err := e.varManager.Substitute(req.ServerAddr); err != nil {
		e.logger.Error("Failed to substitute server_addr", "server_addr", req.ServerAddr, "error", err)
		return nil, fmt.Errorf("failed to substitute server_addr: %w", err)
	} else {
		substituted.ServerAddr = substitutedServerAddr
		e.logger.Debug("ServerAddr substitution result", "original", req.ServerAddr, "substituted", substitutedServerAddr)
	}

	// Substitute gRPC metadata (similar to HTTP headers)
	for key, value := range req.Metadata {
		if substitutedValue, err := e.varManager.Substitute(value); err != nil {
			e.logger.Error("Failed to substitute metadata", "key", key, "value", value, "error", err)
			return nil, fmt.Errorf("failed to substitute metadata %s: %w", key, err)
		} else {
			substituted.Metadata[key] = substitutedValue
			e.logger.Debug("Metadata substitution result", "key", key, "original", value, "substituted", substitutedValue)
		}
	}

	// Substitute gRPC data
	if req.Data != nil {
		switch data := req.Data.(type) {
		case string:
			if substitutedData, err := e.varManager.Substitute(data); err != nil {
				e.logger.Error("Failed to substitute data", "data", data, "error", err)
				return nil, fmt.Errorf("failed to substitute data: %w", err)
			} else {
				substituted.Data = substitutedData
				e.logger.Debug("Data substitution result", "original", data, "substituted", substitutedData)
			}
		case map[string]interface{}:
			if substitutedData, err := e.varManager.SubstituteMap(data); err != nil {
				e.logger.Error("Failed to substitute data map", "data", data, "error", err)
				return nil, fmt.Errorf("failed to substitute data: %w", err)
			} else {
				substituted.Data = substitutedData
				e.logger.Debug("Data map substitution result", "original", data, "substituted", substitutedData)
			}
		default:
			substituted.Data = req.Data
		}
	}

	// Substitute timeout if it contains variables
	if req.Timeout != "" {
		if substitutedTimeout, err := e.varManager.Substitute(req.Timeout); err != nil {
			e.logger.Error("Failed to substitute timeout", "timeout", req.Timeout, "error", err)
			return nil, fmt.Errorf("failed to substitute timeout: %w", err)
		} else {
			substituted.Timeout = substitutedTimeout
			e.logger.Debug("Timeout substitution result", "original", req.Timeout, "substituted", substitutedTimeout)
		}
	}

	e.logger.Debug("Final substituted request", "url", substituted.URL, "method", substituted.Method, "timeout", substituted.Timeout)
	return substituted, nil
}

// captureValues captures values from the response
func (e *Executor) captureValues(response *httpclient.Response, captures map[string]string, result *TestResult) error {
	jsonData, err := response.GetJSONBody()
	if err != nil {
		return fmt.Errorf("failed to parse response as JSON: %w", err)
	}

	for captureKey, jsonPath := range captures {
		// Use validator's ExtractJSONValue which supports advanced JSONPath with filters
		value, err := e.validator.ExtractJSONValue(jsonData, jsonPath)
		if err != nil {
			e.logger.Warn("Failed to capture value", "key", captureKey, "path", jsonPath, "error", err)
			continue
		}

		result.CapturedData[captureKey] = value
		e.varManager.Set(captureKey, value)
		e.logger.Debug("Captured value", "key", captureKey, "value", value)
	}

	return nil
}

// logAPIResponseOnFailure logs the API response when a step fails and verbose mode is enabled
func (e *Executor) logAPIResponseOnFailure(protocol string, httpResponse *httpclient.Response, grpcResponse *grpcclient.Response, dbResponse *dbclient.Response, stepName string) {
	if protocol == "grpc" {
		if grpcResponse != nil {
			jsonData, err := json.MarshalIndent(grpcResponse.Data, "", "  ")
			if err == nil {
				e.logger.Error("API Response (gRPC) on failure",
					"step", stepName,
					"response", string(jsonData))
			} else {
				e.logger.Error("API Response (gRPC) on failure - failed to marshal",
					"step", stepName,
					"error", err)
			}
		}
	} else if protocol == "db" {
		if dbResponse != nil {
			jsonData, err := json.MarshalIndent(dbResponse.Data, "", "  ")
			if err == nil {
				e.logger.Error("API Response (DB) on failure",
					"step", stepName,
					"rows", dbResponse.Rows,
					"response", string(jsonData))
			} else {
				e.logger.Error("API Response (DB) on failure - failed to marshal",
					"step", stepName,
					"error", err)
			}
		}
	} else {
		if httpResponse != nil && len(httpResponse.Body) > 0 {
			// Try to format as JSON if possible, otherwise show as text
			var responseText string
			var jsonData interface{}
			if err := json.Unmarshal(httpResponse.Body, &jsonData); err == nil {
				// Valid JSON, format it nicely
				formatted, err := json.MarshalIndent(jsonData, "", "  ")
				if err == nil {
					responseText = string(formatted)
				} else {
					responseText = string(httpResponse.Body)
				}
			} else {
				// Not JSON, show as text
				responseText = string(httpResponse.Body)
			}
			e.logger.Error("API Response (HTTP) on failure",
				"step", stepName,
				"status_code", httpResponse.StatusCode,
				"response", responseText)
		}
	}
}

// parseTimeout parses a timeout string into duration
func (e *Executor) parseTimeout(timeoutStr string) time.Duration {
	if timeoutStr == "" {
		return e.config.Timeout
	}

	if duration, err := time.ParseDuration(timeoutStr); err == nil {
		return duration
	}

	return e.config.Timeout
}
