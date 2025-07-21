package workflow

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cjp2600/stepwise/internal/config"
	grpcclient "github.com/cjp2600/stepwise/internal/grpc"
	httpclient "github.com/cjp2600/stepwise/internal/http"
	"github.com/cjp2600/stepwise/internal/logger"
	"github.com/cjp2600/stepwise/internal/validation"
	"github.com/cjp2600/stepwise/internal/variables"
	"gopkg.in/yaml.v3"
)

// Workflow represents a complete test workflow
type Workflow struct {
	Name        string                 `yaml:"name" json:"name"`
	Version     string                 `yaml:"version" json:"version"`
	Description string                 `yaml:"description" json:"description"`
	Variables   map[string]interface{} `yaml:"variables" json:"variables"`
	Steps       []Step                 `yaml:"steps" json:"steps"`
	Groups      []StepGroup            `yaml:"groups" json:"groups"`
}

// StepGroup represents a group of steps that can be executed together
type StepGroup struct {
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description" json:"description"`
	Parallel    bool   `yaml:"parallel" json:"parallel"`
	Condition   string `yaml:"condition" json:"condition"`
	Steps       []Step `yaml:"steps" json:"steps"`
}

// Step represents a single test step
type Step struct {
	Name        string                      `yaml:"name" json:"name"`
	Description string                      `yaml:"description" json:"description"`
	Request     Request                     `yaml:"request" json:"request"`
	Validate    []validation.ValidationRule `yaml:"validate" json:"validate"`
	Capture     map[string]string           `yaml:"capture" json:"capture"`
	Condition   string                      `yaml:"condition" json:"condition"`
	Retry       int                         `yaml:"retry" json:"retry"`
	RetryDelay  string                      `yaml:"retry_delay" json:"retry_delay"`
	Timeout     string                      `yaml:"timeout" json:"timeout"`
}

// Request represents an HTTP or gRPC request
type Request struct {
	// Protocol type: "http" or "grpc"
	Protocol string `yaml:"protocol" json:"protocol"`

	// HTTP fields
	Method  string            `yaml:"method" json:"method"`
	URL     string            `yaml:"url" json:"url"`
	Headers map[string]string `yaml:"headers" json:"headers"`
	Body    interface{}       `yaml:"body" json:"body"`
	Query   map[string]string `yaml:"query" json:"query"`
	Auth    *httpclient.Auth  `yaml:"auth" json:"auth"`

	// gRPC fields
	Service    string            `yaml:"service" json:"service"`
	GRPCMethod string            `yaml:"grpc_method" json:"grpc_method"`
	Data       interface{}       `yaml:"data" json:"data"`
	Metadata   map[string]string `yaml:"metadata" json:"metadata"`
	ServerAddr string            `yaml:"server_addr" json:"server_addr"`
	Insecure   bool              `yaml:"insecure" json:"insecure"`

	// Common fields
	Timeout string `yaml:"timeout" json:"timeout"`
}

// TestResult represents the result of a test step
type TestResult struct {
	Name         string                        `json:"name"`
	Status       string                        `json:"status"`
	Duration     time.Duration                 `json:"duration"`
	Error        string                        `json:"error,omitempty"`
	Validations  []validation.ValidationResult `json:"validations,omitempty"`
	CapturedData map[string]interface{}        `json:"captured_data,omitempty"`
	Retries      int                           `json:"retries,omitempty"`
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

// Executor executes workflows
type Executor struct {
	config     *config.Config
	logger     *logger.Logger
	httpClient *httpclient.Client
	grpcClient *grpcclient.Client
	validator  *validation.Validator
	varManager *variables.Manager
}

// NewExecutor creates a new workflow executor
func NewExecutor(cfg *config.Config, log *logger.Logger) *Executor {
	return &Executor{
		config:     cfg,
		logger:     log,
		httpClient: httpclient.NewClient(cfg.Timeout, log),
		grpcClient: nil, // Will be initialized when needed
		validator:  validation.NewValidator(log),
		varManager: variables.NewManager(log),
	}
}

// Load loads a workflow from a file
func Load(filename string) (*Workflow, error) {
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

	return &workflow, nil
}

// Execute executes a workflow and returns the results
func (e *Executor) Execute(wf *Workflow) ([]TestResult, error) {
	e.logger.Info("Starting workflow execution", "name", wf.Name)
	startTime := time.Now()

	// Initialize variables
	e.initializeVariables(wf.Variables)

	var allResults []TestResult

	// Execute individual steps
	for _, step := range wf.Steps {
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
				allResults = append(allResults, *result)
				continue
			}
		}

		if err := e.executeStep(&step, result); err != nil {
			result.Status = "failed"
			result.Error = err.Error()
			e.logger.Error("Step execution failed", "step", step.Name, "error", err)
		} else {
			result.Status = "passed"
		}

		allResults = append(allResults, *result)
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

		if err := e.executeStep(&step, result); err != nil {
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

			if err := e.executeStep(&step, result); err != nil {
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
	startTime := time.Now()
	maxRetries := step.Retry
	if maxRetries == 0 {
		maxRetries = 1 // Default to 1 attempt
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
		} else {
			// Execute HTTP request (default)
			httpReq := &httpclient.Request{
				Method:  substitutedReq.Method,
				URL:     substitutedReq.URL,
				Headers: substitutedReq.Headers,
				Body:    substitutedReq.Body,
				Query:   substitutedReq.Query,
				Timeout: e.parseTimeout(substitutedReq.Timeout),
				Auth:    substitutedReq.Auth,
			}
			httpResponse, requestErr = e.httpClient.Execute(httpReq)
		}

		if requestErr != nil {
			lastError = fmt.Errorf("request failed: %w", requestErr)
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
			} else {
				validationResults, validationErr = e.validator.Validate(httpResponse, step.Validate)
			}

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

// initializeVariables initializes the variable manager with workflow variables
func (e *Executor) initializeVariables(vars map[string]interface{}) {
	for key, value := range vars {
		e.varManager.Set(key, value)
	}
}

// substituteRequestVariables substitutes variables in the request
func (e *Executor) substituteRequestVariables(req *Request) (*Request, error) {
	substituted := &Request{
		Protocol:   req.Protocol,
		Method:     req.Method,
		URL:        req.URL,
		Headers:    make(map[string]string),
		Body:       req.Body,
		Query:      make(map[string]string),
		Service:    req.Service,
		GRPCMethod: req.GRPCMethod,
		Data:       req.Data,
		Metadata:   req.Metadata,
		ServerAddr: req.ServerAddr,
		Insecure:   req.Insecure,
		Timeout:    req.Timeout,
	}

	// Substitute URL
	if substitutedURL, err := e.varManager.Substitute(req.URL); err != nil {
		return nil, fmt.Errorf("failed to substitute URL: %w", err)
	} else {
		substituted.URL = substitutedURL
	}

	// Substitute headers
	for key, value := range req.Headers {
		if substitutedValue, err := e.varManager.Substitute(value); err != nil {
			return nil, fmt.Errorf("failed to substitute header %s: %w", key, err)
		} else {
			substituted.Headers[key] = substitutedValue
		}
	}

	// Substitute query parameters
	for key, value := range req.Query {
		if substitutedValue, err := e.varManager.Substitute(value); err != nil {
			return nil, fmt.Errorf("failed to substitute query %s: %w", key, err)
		} else {
			substituted.Query[key] = substitutedValue
		}
	}

	// Substitute body
	if req.Body != nil {
		switch body := req.Body.(type) {
		case string:
			if substitutedBody, err := e.varManager.Substitute(body); err != nil {
				return nil, fmt.Errorf("failed to substitute body: %w", err)
			} else {
				substituted.Body = substitutedBody
			}
		case map[string]interface{}:
			if substitutedBody, err := e.varManager.SubstituteMap(body); err != nil {
				return nil, fmt.Errorf("failed to substitute body: %w", err)
			} else {
				substituted.Body = substitutedBody
			}
		default:
			substituted.Body = req.Body
		}
	}

	// Substitute gRPC fields
	if substitutedServerAddr, err := e.varManager.Substitute(req.ServerAddr); err != nil {
		return nil, fmt.Errorf("failed to substitute server_addr: %w", err)
	} else {
		substituted.ServerAddr = substitutedServerAddr
	}

	// Substitute gRPC data
	if req.Data != nil {
		switch data := req.Data.(type) {
		case string:
			if substitutedData, err := e.varManager.Substitute(data); err != nil {
				return nil, fmt.Errorf("failed to substitute data: %w", err)
			} else {
				substituted.Data = substitutedData
			}
		case map[string]interface{}:
			if substitutedData, err := e.varManager.SubstituteMap(data); err != nil {
				return nil, fmt.Errorf("failed to substitute data: %w", err)
			} else {
				substituted.Data = substitutedData
			}
		default:
			substituted.Data = req.Data
		}
	}

	return substituted, nil
}

// captureValues captures values from the response
func (e *Executor) captureValues(response *httpclient.Response, captures map[string]string, result *TestResult) error {
	jsonData, err := response.GetJSONBody()
	if err != nil {
		return fmt.Errorf("failed to parse response as JSON: %w", err)
	}

	for captureKey, jsonPath := range captures {
		value, err := e.extractJSONValue(jsonData, jsonPath)
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

// extractJSONValue extracts a value from JSON data using JSONPath-like syntax
func (e *Executor) extractJSONValue(data interface{}, path string) (interface{}, error) {
	// Simple JSONPath-like extraction
	// For now, handle basic cases like "$.key" or "$[0]"
	if path == "$" {
		return data, nil
	}

	if strings.HasPrefix(path, "$.") {
		key := strings.TrimPrefix(path, "$.")
		if mapData, ok := data.(map[string]interface{}); ok {
			if value, exists := mapData[key]; exists {
				return value, nil
			}
		}
		return nil, fmt.Errorf("key not found: %s", key)
	}

	if strings.HasPrefix(path, "$[") && strings.HasSuffix(path, "]") {
		indexStr := strings.TrimPrefix(strings.TrimSuffix(path, "]"), "$[")
		index, err := strconv.Atoi(indexStr)
		if err != nil {
			return nil, fmt.Errorf("invalid array index: %s", indexStr)
		}
		if arrayData, ok := data.([]interface{}); ok {
			if index >= 0 && index < len(arrayData) {
				return arrayData[index], nil
			}
		}
		return nil, fmt.Errorf("array index out of bounds: %d", index)
	}

	return nil, fmt.Errorf("unsupported JSON path: %s", path)
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
