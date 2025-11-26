package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"sync"
	"time"

	"github.com/cjp2600/stepwise/internal/logger"
)

// Client represents an MCP client for making requests
type Client struct {
	transport     Transport
	logger        *logger.Logger
	initialized   bool
	serverInfo    *ServerInfo
	clientInfo    *ClientInfo
	requestID     int64
	requestIDLock sync.Mutex
	httpClient    *http.Client
}

// Transport represents the transport mechanism for MCP communication
type Transport interface {
	SendRequest(ctx context.Context, req *JSONRPCRequest) (*JSONRPCResponse, error)
	Close() error
}

// ServerInfo represents MCP server information
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// ClientInfo represents MCP client information
type ClientInfo struct {
	Name         string                 `json:"name"`
	Version      string                 `json:"version"`
	Capabilities map[string]interface{} `json:"capabilities,omitempty"`
}

// JSONRPCRequest represents a JSON-RPC 2.0 request
type JSONRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response
type JSONRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *JSONRPCError   `json:"error,omitempty"`
}

// JSONRPCError represents a JSON-RPC error
type JSONRPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Request represents an MCP request
type Request struct {
	// Transport type: "stdio", "http", "websocket"
	Transport string `yaml:"transport" json:"transport"`

	// For stdio transport
	Command string   `yaml:"command" json:"command"`
	Args    []string `yaml:"args,omitempty" json:"args,omitempty"`

	// For HTTP/WebSocket transport
	URL     string            `yaml:"url" json:"url"`
	Headers map[string]string `yaml:"headers,omitempty" json:"headers,omitempty"`

	// MCP method to call
	Method string `yaml:"method" json:"method"`

	// Method-specific parameters
	Params map[string]interface{} `yaml:"params,omitempty" json:"params,omitempty"`

	// Timeout for the request
	Timeout string `yaml:"timeout" json:"timeout"`

	// Client info for initialization
	ClientInfo *ClientInfo `yaml:"client_info,omitempty" json:"client_info,omitempty"`
}

// Response represents an MCP response
type Response struct {
	Result   interface{}   `json:"result"`
	Error    *JSONRPCError `json:"error,omitempty"`
	Duration time.Duration `json:"duration"`
	Method   string        `json:"method"`
}

// Tool represents an MCP tool
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

// Resource represents an MCP resource
type Resource struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	MimeType    string `json:"mimeType,omitempty"`
}

// Prompt represents an MCP prompt
type Prompt struct {
	Name        string           `json:"name"`
	Description string           `json:"description,omitempty"`
	Arguments   []PromptArgument `json:"arguments,omitempty"`
}

// PromptArgument represents a prompt argument
type PromptArgument struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required,omitempty"`
}

// InitializeParams represents parameters for initialize method
type InitializeParams struct {
	ProtocolVersion string                 `json:"protocolVersion"`
	Capabilities    map[string]interface{} `json:"capabilities,omitempty"`
	ClientInfo      *ClientInfo            `json:"clientInfo"`
}

// InitializeResult represents result from initialize method
type InitializeResult struct {
	ProtocolVersion string                 `json:"protocolVersion"`
	Capabilities    map[string]interface{} `json:"capabilities,omitempty"`
	ServerInfo      *ServerInfo            `json:"serverInfo"`
}

// ToolsListResult represents result from tools/list method
type ToolsListResult struct {
	Tools []Tool `json:"tools"`
}

// ToolsCallParams represents parameters for tools/call method
type ToolsCallParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

// ToolsCallResult represents result from tools/call method
type ToolsCallResult struct {
	Content []Content `json:"content"`
	IsError bool      `json:"isError,omitempty"`
}

// Content represents content in a tool call result
type Content struct {
	Type string      `json:"type"`
	Text string      `json:"text,omitempty"`
	Data interface{} `json:"data,omitempty"`
}

// ResourcesListResult represents result from resources/list method
type ResourcesListResult struct {
	Resources []Resource `json:"resources"`
}

// ResourcesReadParams represents parameters for resources/read method
type ResourcesReadParams struct {
	URI string `json:"uri"`
}

// ResourcesReadResult represents result from resources/read method
type ResourcesReadResult struct {
	Contents []Content `json:"contents"`
}

// PromptsListResult represents result from prompts/list method
type PromptsListResult struct {
	Prompts []Prompt `json:"prompts"`
}

// PromptsGetParams represents parameters for prompts/get method
type PromptsGetParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

// PromptsGetResult represents result from prompts/get method
type PromptsGetResult struct {
	Description string    `json:"description,omitempty"`
	Messages    []Message `json:"messages"`
}

// Message represents a message in a prompt
type Message struct {
	Role    string                 `json:"role"`
	Content interface{}            `json:"content"`
	Data    map[string]interface{} `json:"data,omitempty"`
}

// NewClient creates a new MCP client
func NewClient(req *Request, log *logger.Logger) (*Client, error) {
	client := &Client{
		logger:      log,
		initialized: false,
		requestID:   0,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	// Set default client info
	if req.ClientInfo == nil {
		client.clientInfo = &ClientInfo{
			Name:    "stepwise",
			Version: "1.0.0",
			Capabilities: map[string]interface{}{
				"tools":     map[string]interface{}{},
				"resources": map[string]interface{}{},
				"prompts":   map[string]interface{}{},
			},
		}
	} else {
		client.clientInfo = req.ClientInfo
	}

	// Create transport based on type
	var err error
	switch req.Transport {
	case "stdio":
		client.transport, err = NewStdioTransport(req.Command, req.Args, log)
	case "http", "https":
		client.transport, err = NewHTTPTransport(req.URL, req.Headers, log)
	case "websocket", "ws", "wss":
		// WebSocket transport would be implemented here
		return nil, fmt.Errorf("websocket transport not yet implemented")
	default:
		return nil, fmt.Errorf("unsupported transport type: %s", req.Transport)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create transport: %w", err)
	}

	// Initialize the connection
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := client.initialize(ctx); err != nil {
		client.transport.Close()
		return nil, fmt.Errorf("failed to initialize MCP connection: %w", err)
	}

	return client, nil
}

// initialize initializes the MCP connection
func (c *Client) initialize(ctx context.Context) error {
	params := InitializeParams{
		ProtocolVersion: "2024-11-05",
		Capabilities: map[string]interface{}{
			"tools":     map[string]interface{}{},
			"resources": map[string]interface{}{},
			"prompts":   map[string]interface{}{},
		},
		ClientInfo: c.clientInfo,
	}

	resp, err := c.CallMethod(ctx, "initialize", params)
	if err != nil {
		return fmt.Errorf("initialize failed: %w", err)
	}

	var result InitializeResult
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return fmt.Errorf("failed to unmarshal initialize result: %w", err)
	}

	c.serverInfo = result.ServerInfo
	c.initialized = true

	c.logger.Info("MCP connection initialized",
		"server_name", result.ServerInfo.Name,
		"server_version", result.ServerInfo.Version,
		"protocol_version", result.ProtocolVersion)

	// Send initialized notification
	notification := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "notifications/initialized",
	}
	// Notifications don't have ID and don't expect response
	_, _ = c.transport.SendRequest(ctx, &notification)

	return nil
}

// CallMethod calls an MCP method
func (c *Client) CallMethod(ctx context.Context, method string, params interface{}) (*JSONRPCResponse, error) {
	if !c.initialized && method != "initialize" {
		return nil, fmt.Errorf("client not initialized, call initialize first")
	}

	c.requestIDLock.Lock()
	c.requestID++
	id := c.requestID
	c.requestIDLock.Unlock()

	req := &JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}

	return c.transport.SendRequest(ctx, req)
}

// Execute executes an MCP request
func (c *Client) Execute(req *Request) (*Response, error) {
	start := time.Now()

	// Parse timeout
	timeout := 30 * time.Second
	if req.Timeout != "" {
		if duration, err := time.ParseDuration(req.Timeout); err == nil {
			timeout = duration
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	c.logger.Debug("Making MCP request",
		"method", req.Method,
		"transport", req.Transport)

	// Call the method
	var params interface{}
	if len(req.Params) > 0 {
		params = req.Params
	}

	resp, err := c.CallMethod(ctx, req.Method, params)
	if err != nil {
		return nil, fmt.Errorf("MCP request failed: %w", err)
	}

	// Parse result based on method
	var result interface{}
	if resp.Error != nil {
		return &Response{
			Error:    resp.Error,
			Duration: time.Since(start),
			Method:   req.Method,
		}, nil
	}

	// Unmarshal result based on method type
	switch req.Method {
	case "tools/list":
		var toolsResult ToolsListResult
		if err := json.Unmarshal(resp.Result, &toolsResult); err == nil {
			result = toolsResult
		} else {
			result = resp.Result
		}
	case "tools/call":
		var callResult ToolsCallResult
		if err := json.Unmarshal(resp.Result, &callResult); err == nil {
			result = callResult
		} else {
			result = resp.Result
		}
	case "resources/list":
		var resourcesResult ResourcesListResult
		if err := json.Unmarshal(resp.Result, &resourcesResult); err == nil {
			result = resourcesResult
		} else {
			result = resp.Result
		}
	case "resources/read":
		var readResult ResourcesReadResult
		if err := json.Unmarshal(resp.Result, &readResult); err == nil {
			result = readResult
		} else {
			result = resp.Result
		}
	case "prompts/list":
		var promptsResult PromptsListResult
		if err := json.Unmarshal(resp.Result, &promptsResult); err == nil {
			result = promptsResult
		} else {
			result = resp.Result
		}
	case "prompts/get":
		var getResult PromptsGetResult
		if err := json.Unmarshal(resp.Result, &getResult); err == nil {
			result = getResult
		} else {
			result = resp.Result
		}
	default:
		// For other methods, return raw result
		if err := json.Unmarshal(resp.Result, &result); err != nil {
			result = resp.Result
		}
	}

	duration := time.Since(start)

	c.logger.Debug("Received MCP response",
		"method", req.Method,
		"duration", duration)

	return &Response{
		Result:   result,
		Error:    resp.Error,
		Duration: duration,
		Method:   req.Method,
	}, nil
}

// Close closes the MCP connection
func (c *Client) Close() error {
	if c.transport != nil {
		return c.transport.Close()
	}
	return nil
}

// StdioTransport implements Transport using stdio
type StdioTransport struct {
	cmd     *exec.Cmd
	stdin   io.WriteCloser
	stdout  io.ReadCloser
	logger  *logger.Logger
	decoder *json.Decoder
	encoder *json.Encoder
	lock    sync.Mutex
}

// NewStdioTransport creates a new stdio transport
func NewStdioTransport(command string, args []string, log *logger.Logger) (*StdioTransport, error) {
	cmd := exec.Command(command, args...)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		stdin.Close()
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		stdin.Close()
		stdout.Close()
		return nil, fmt.Errorf("failed to start command: %w", err)
	}

	transport := &StdioTransport{
		cmd:     cmd,
		stdin:   stdin,
		stdout:  stdout,
		logger:  log,
		decoder: json.NewDecoder(stdout),
		encoder: json.NewEncoder(stdin),
	}

	return transport, nil
}

// SendRequest sends a JSON-RPC request via stdio
func (t *StdioTransport) SendRequest(ctx context.Context, req *JSONRPCRequest) (*JSONRPCResponse, error) {
	t.lock.Lock()
	defer t.lock.Unlock()

	// Encode and send request
	if err := t.encoder.Encode(req); err != nil {
		return nil, fmt.Errorf("failed to encode request: %w", err)
	}

	// For notifications (no ID), don't wait for response
	if req.ID == nil {
		return &JSONRPCResponse{JSONRPC: "2.0"}, nil
	}

	// Read and decode response
	var resp JSONRPCResponse
	if err := t.decoder.Decode(&resp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Verify response ID matches request ID
	if resp.ID != req.ID {
		return nil, fmt.Errorf("response ID mismatch: expected %v, got %v", req.ID, resp.ID)
	}

	return &resp, nil
}

// Close closes the stdio transport
func (t *StdioTransport) Close() error {
	if t.stdin != nil {
		t.stdin.Close()
	}
	if t.stdout != nil {
		t.stdout.Close()
	}
	if t.cmd != nil && t.cmd.Process != nil {
		return t.cmd.Process.Kill()
	}
	return nil
}

// HTTPTransport implements Transport using HTTP
type HTTPTransport struct {
	url     string
	headers map[string]string
	logger  *logger.Logger
	client  *http.Client
}

// NewHTTPTransport creates a new HTTP transport
func NewHTTPTransport(url string, headers map[string]string, log *logger.Logger) (*HTTPTransport, error) {
	return &HTTPTransport{
		url:     url,
		headers: headers,
		logger:  log,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// SendRequest sends a JSON-RPC request via HTTP
func (t *HTTPTransport) SendRequest(ctx context.Context, req *JSONRPCRequest) (*JSONRPCResponse, error) {
	// Marshal request
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", t.url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	for key, value := range t.headers {
		httpReq.Header.Set(key, value)
	}

	// Send request
	resp, err := t.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Decode response
	var jsonrpcResp JSONRPCResponse
	if err := json.NewDecoder(resp.Body).Decode(&jsonrpcResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// For notifications (no ID), return empty response
	if req.ID == nil {
		return &JSONRPCResponse{JSONRPC: "2.0"}, nil
	}

	return &jsonrpcResp, nil
}

// Close closes the HTTP transport (no-op for HTTP)
func (t *HTTPTransport) Close() error {
	return nil
}

// Helper methods for common MCP operations

// ListTools lists available tools
func (c *Client) ListTools(ctx context.Context) ([]Tool, error) {
	resp, err := c.CallMethod(ctx, "tools/list", nil)
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("tools/list error: %s", resp.Error.Message)
	}

	var result ToolsListResult
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tools list: %w", err)
	}

	return result.Tools, nil
}

// CallTool calls a tool
func (c *Client) CallTool(ctx context.Context, name string, arguments map[string]interface{}) (*ToolsCallResult, error) {
	params := ToolsCallParams{
		Name:      name,
		Arguments: arguments,
	}

	resp, err := c.CallMethod(ctx, "tools/call", params)
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("tools/call error: %s", resp.Error.Message)
	}

	var result ToolsCallResult
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tool call result: %w", err)
	}

	return &result, nil
}

// ListResources lists available resources
func (c *Client) ListResources(ctx context.Context) ([]Resource, error) {
	resp, err := c.CallMethod(ctx, "resources/list", nil)
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("resources/list error: %s", resp.Error.Message)
	}

	var result ResourcesListResult
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal resources list: %w", err)
	}

	return result.Resources, nil
}

// ReadResource reads a resource
func (c *Client) ReadResource(ctx context.Context, uri string) (*ResourcesReadResult, error) {
	params := ResourcesReadParams{
		URI: uri,
	}

	resp, err := c.CallMethod(ctx, "resources/read", params)
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("resources/read error: %s", resp.Error.Message)
	}

	var result ResourcesReadResult
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal resource read result: %w", err)
	}

	return &result, nil
}

// ListPrompts lists available prompts
func (c *Client) ListPrompts(ctx context.Context) ([]Prompt, error) {
	resp, err := c.CallMethod(ctx, "prompts/list", nil)
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("prompts/list error: %s", resp.Error.Message)
	}

	var result PromptsListResult
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal prompts list: %w", err)
	}

	return result.Prompts, nil
}

// GetPrompt gets a prompt
func (c *Client) GetPrompt(ctx context.Context, name string, arguments map[string]interface{}) (*PromptsGetResult, error) {
	params := PromptsGetParams{
		Name:      name,
		Arguments: arguments,
	}

	resp, err := c.CallMethod(ctx, "prompts/get", params)
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("prompts/get error: %s", resp.Error.Message)
	}

	var result PromptsGetResult
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal prompt get result: %w", err)
	}

	return &result, nil
}
