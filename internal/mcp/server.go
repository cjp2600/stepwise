package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/cjp2600/stepwise/internal/config"
	"github.com/cjp2600/stepwise/internal/logger"
)

// MCPOutputHandler interface for MCP output
type MCPOutputHandler interface {
	SendProgress(update ProgressUpdate) error
	SendLog(level, message string, fields map[string]interface{}) error
	SendResult(result interface{}) error
	SendOutput(text string) error
}

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

// CLIAppRunner interface for running CLI commands
type CLIAppRunner interface {
	RunWithMCP(args []string, outputHandler MCPOutputHandler) error
}

// Server represents an MCP server that handles JSON-RPC requests via stdin/stdout
type Server struct {
	decoder    *json.Decoder
	encoder    *json.Encoder
	requestID  int64
	idLock     sync.Mutex
	config     *config.Config
	logger     *logger.Logger
	cliApp     CLIAppRunner
	outputLock sync.Mutex
}

// NewServer creates a new MCP server
func NewServer(cfg *config.Config, log *logger.Logger, cliApp CLIAppRunner) *Server {
	return &Server{
		decoder:   json.NewDecoder(os.Stdin),
		encoder:   json.NewEncoder(os.Stdout),
		requestID: 0,
		config:    cfg,
		logger:    log,
		cliApp:    cliApp,
	}
}

// Run starts the MCP server and processes requests
func (s *Server) Run(ctx context.Context) error {
	// Wait for initialize request from client
	// The client sends initialize, we respond with server info

	// Wait for initialize request from client
	var initReq JSONRPCRequest
	if err := s.decoder.Decode(&initReq); err != nil {
		return fmt.Errorf("failed to read initialize request: %w", err)
	}

	if initReq.Method != "initialize" {
		return fmt.Errorf("expected initialize request, got %s", initReq.Method)
	}

	// Send initialize response
	initResult := InitializeResult{
		ProtocolVersion: "2024-11-05",
		Capabilities: map[string]interface{}{
			"tools": map[string]interface{}{
				"listChanged": true,
			},
		},
		ServerInfo: &ServerInfo{
			Name:    "stepwise",
			Version: "1.0.0",
		},
	}
	initResp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      initReq.ID,
		Result:  mustMarshal(initResult),
	}
	if err := s.encoder.Encode(initResp); err != nil {
		return fmt.Errorf("failed to send initialize response: %w", err)
	}

	// Wait for initialized notification
	var initializedNotification JSONRPCRequest
	if err := s.decoder.Decode(&initializedNotification); err != nil {
		return fmt.Errorf("failed to read initialized notification: %w", err)
	}

	if initializedNotification.Method != "notifications/initialized" {
		return fmt.Errorf("expected initialized notification, got %s", initializedNotification.Method)
	}

	// Now process requests
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			var req JSONRPCRequest
			if err := s.decoder.Decode(&req); err != nil {
				if err == io.EOF {
					return nil
				}
				s.sendError(nil, -32700, "Parse error", err.Error())
				continue
			}

			// Handle the request synchronously to maintain order
			s.handleRequest(&req)
		}
	}
}

// handleRequest handles a single JSON-RPC request
func (s *Server) handleRequest(req *JSONRPCRequest) {
	var resp *JSONRPCResponse

	switch req.Method {
	case "tools/list":
		resp = s.handleToolsList(req)
	case "tools/call":
		resp = s.handleToolsCall(req)
	case "ping":
		resp = &JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  mustMarshal(map[string]string{"status": "pong"}),
		}
	default:
		resp = &JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &JSONRPCError{
				Code:    -32601,
				Message: "Method not found",
				Data:    req.Method,
			},
		}
	}

	if resp != nil {
		s.outputLock.Lock()
		s.encoder.Encode(resp)
		s.outputLock.Unlock()
	}
}

// handleToolsList handles the tools/list method
func (s *Server) handleToolsList(req *JSONRPCRequest) *JSONRPCResponse {
	tools := []Tool{
		{
			Name:        "stepwise_run",
			Description: "Run a workflow file or directory",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": "Path to workflow file or directory",
					},
					"parallel": map[string]interface{}{
						"type":        "integer",
						"description": "Number of parallel workflow executions",
						"default":     1,
					},
					"recursive": map[string]interface{}{
						"type":        "boolean",
						"description": "Search recursively in subdirectories",
						"default":     false,
					},
					"verbose": map[string]interface{}{
						"type":        "boolean",
						"description": "Enable verbose logging",
						"default":     false,
					},
					"fail_fast": map[string]interface{}{
						"type":        "boolean",
						"description": "Stop execution on first test failure",
						"default":     false,
					},
					"html_report": map[string]interface{}{
						"type":        "boolean",
						"description": "Generate HTML report",
						"default":     false,
					},
					"html_report_path": map[string]interface{}{
						"type":        "string",
						"description": "Path for HTML report file",
					},
				},
				"required": []string{"path"},
			},
		},
		{
			Name:        "stepwise_validate",
			Description: "Validate a workflow file",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": "Path to workflow file",
					},
				},
				"required": []string{"path"},
			},
		},
		{
			Name:        "stepwise_info",
			Description: "Show workflow information",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": "Path to workflow file",
					},
				},
				"required": []string{"path"},
			},
		},
	}

	return &JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  mustMarshal(ToolsListResult{Tools: tools}),
	}
}

// handleToolsCall handles the tools/call method
func (s *Server) handleToolsCall(req *JSONRPCRequest) *JSONRPCResponse {
	var params ToolsCallParams
	paramsBytes, _ := json.Marshal(req.Params)
	if err := json.Unmarshal(paramsBytes, &params); err != nil {
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &JSONRPCError{
				Code:    -32602,
				Message: "Invalid params",
				Data:    err.Error(),
			},
		}
	}

	// Create MCP output handler
	var outputHandler MCPOutputHandler = NewMCPOutputHandler(s.encoder, &s.outputLock)

	// Execute the tool based on name
	var result *ToolsCallResult
	var err error

	switch params.Name {
	case "stepwise_run":
		result, err = s.handleRunTool(params.Arguments, outputHandler)
	case "stepwise_validate":
		result, err = s.handleValidateTool(params.Arguments, outputHandler)
	case "stepwise_info":
		result, err = s.handleInfoTool(params.Arguments, outputHandler)
	default:
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &JSONRPCError{
				Code:    -32601,
				Message: "Tool not found",
				Data:    params.Name,
			},
		}
	}

	if err != nil {
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &JSONRPCError{
				Code:    -32000,
				Message: "Execution error",
				Data:    err.Error(),
			},
		}
	}

	return &JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  mustMarshal(result),
	}
}

// handleRunTool handles the stepwise_run tool
func (s *Server) handleRunTool(args map[string]interface{}, outputHandler MCPOutputHandler) (*ToolsCallResult, error) {
	path, ok := args["path"].(string)
	if !ok {
		return nil, fmt.Errorf("path parameter is required")
	}

	// Build command arguments
	cmdArgs := []string{"run", path}

	if parallel, ok := args["parallel"].(float64); ok {
		cmdArgs = append(cmdArgs, "--parallel", fmt.Sprintf("%.0f", parallel))
	}
	if recursive, ok := args["recursive"].(bool); ok && recursive {
		cmdArgs = append(cmdArgs, "--recursive")
	}
	if verbose, ok := args["verbose"].(bool); ok && verbose {
		cmdArgs = append(cmdArgs, "--verbose")
	}
	if failFast, ok := args["fail_fast"].(bool); ok && failFast {
		cmdArgs = append(cmdArgs, "--fail-fast")
	}
	if htmlReport, ok := args["html_report"].(bool); ok && htmlReport {
		cmdArgs = append(cmdArgs, "--html-report")
	}
	if htmlReportPath, ok := args["html_report_path"].(string); ok && htmlReportPath != "" {
		cmdArgs = append(cmdArgs, "--html-report-path", htmlReportPath)
	}

	// Execute with MCP output handler
	err := s.cliApp.RunWithMCP(cmdArgs, outputHandler)

	if err != nil {
		return &ToolsCallResult{
			Content: []Content{
				{
					Type: "text",
					Text: fmt.Sprintf("Error: %v", err),
				},
			},
			IsError: true,
		}, nil
	}

	return &ToolsCallResult{
		Content: []Content{
			{
				Type: "text",
				Text: "Workflow execution completed successfully",
			},
		},
		IsError: false,
	}, nil
}

// handleValidateTool handles the stepwise_validate tool
func (s *Server) handleValidateTool(args map[string]interface{}, outputHandler MCPOutputHandler) (*ToolsCallResult, error) {
	path, ok := args["path"].(string)
	if !ok {
		return nil, fmt.Errorf("path parameter is required")
	}

	cmdArgs := []string{"validate", path}
	err := s.cliApp.RunWithMCP(cmdArgs, outputHandler)

	if err != nil {
		return &ToolsCallResult{
			Content: []Content{
				{
					Type: "text",
					Text: fmt.Sprintf("Validation failed: %v", err),
				},
			},
			IsError: true,
		}, nil
	}

	return &ToolsCallResult{
		Content: []Content{
			{
				Type: "text",
				Text: "Workflow is valid",
			},
		},
		IsError: false,
	}, nil
}

// handleInfoTool handles the stepwise_info tool
func (s *Server) handleInfoTool(args map[string]interface{}, outputHandler MCPOutputHandler) (*ToolsCallResult, error) {
	path, ok := args["path"].(string)
	if !ok {
		return nil, fmt.Errorf("path parameter is required")
	}

	cmdArgs := []string{"info", path}
	err := s.cliApp.RunWithMCP(cmdArgs, outputHandler)

	if err != nil {
		return &ToolsCallResult{
			Content: []Content{
				{
					Type: "text",
					Text: fmt.Sprintf("Error: %v", err),
				},
			},
			IsError: true,
		}, nil
	}

	return &ToolsCallResult{
		Content: []Content{
			{
				Type: "text",
				Text: "Workflow information retrieved",
			},
		},
		IsError: false,
	}, nil
}

// sendError sends an error response
func (s *Server) sendError(id interface{}, code int, message, data string) {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &JSONRPCError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}
	s.outputLock.Lock()
	s.encoder.Encode(resp)
	s.outputLock.Unlock()
}

// mustMarshal marshals data to JSON, panicking on error
func mustMarshal(v interface{}) json.RawMessage {
	data, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal: %v", err))
	}
	return json.RawMessage(data)
}

// MCPOutputHandlerImpl implements MCPOutputHandler interface
type MCPOutputHandlerImpl struct {
	encoder    *json.Encoder
	outputLock *sync.Mutex
}

// NewMCPOutputHandler creates a new MCP output handler
func NewMCPOutputHandler(encoder *json.Encoder, outputLock *sync.Mutex) *MCPOutputHandlerImpl {
	return &MCPOutputHandlerImpl{
		encoder:    encoder,
		outputLock: outputLock,
	}
}

// SendNotification sends a JSON-RPC notification
func (h *MCPOutputHandlerImpl) SendNotification(method string, params interface{}) error {
	notification := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
		// No ID for notifications
	}

	h.outputLock.Lock()
	defer h.outputLock.Unlock()
	return h.encoder.Encode(notification)
}

// SendProgress sends a progress update notification
func (h *MCPOutputHandlerImpl) SendProgress(update ProgressUpdate) error {
	return h.SendNotification("stepwise/progress", map[string]interface{}{
		"step_name":           update.StepName,
		"step_index":          update.StepIndex,
		"total_steps":         update.TotalSteps,
		"status":              update.Status,
		"duration_ms":         update.Duration.Milliseconds(),
		"error":               update.Error,
		"validation_count":   update.ValidationCount,
		"validations_passed":  update.ValidationsPassed,
	})
}

// SendLog sends a log message notification
func (h *MCPOutputHandlerImpl) SendLog(level, message string, fields map[string]interface{}) error {
	params := map[string]interface{}{
		"level":   level,
		"message": message,
	}
	if fields != nil {
		params["fields"] = fields
	}
	return h.SendNotification("stepwise/log", params)
}

// SendResult sends a test result notification
func (h *MCPOutputHandlerImpl) SendResult(result interface{}) error {
	return h.SendNotification("stepwise/result", result)
}

// SendOutput sends a general output notification
func (h *MCPOutputHandlerImpl) SendOutput(text string) error {
	return h.SendNotification("stepwise/output", map[string]interface{}{
		"text": text,
	})
}

