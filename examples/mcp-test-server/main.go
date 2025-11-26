package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

// JSONRPCRequest represents a JSON-RPC 2.0 request
type JSONRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response
type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

// Tool represents an MCP tool
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

// Resource represents an MCP resource
type Resource struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description"`
	MimeType    string `json:"mimeType,omitempty"`
}

// Content represents content in a response
type Content struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/mcp", handleMCP)
	http.HandleFunc("/health", handleHealth)

	fmt.Printf("MCP Test Server starting on port %s\n", port)
	fmt.Printf("MCP endpoint: http://localhost:%s/mcp\n", port)
	fmt.Printf("Health check: http://localhost:%s/health\n", port)

	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func handleMCP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req JSONRPCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, nil, -32700, "Parse error", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Handle notifications (no ID)
	if req.ID == nil {
		if req.Method == "notifications/initialized" {
			w.WriteHeader(http.StatusOK)
			return
		}
	}

	var resp JSONRPCResponse
	resp.JSONRPC = "2.0"
	resp.ID = req.ID

	switch req.Method {
	case "initialize":
		resp.Result = handleInitialize(req.Params)
	case "tools/list":
		resp.Result = handleToolsList()
	case "tools/call":
		resp.Result = handleToolsCall(req.Params)
	case "resources/list":
		resp.Result = handleResourcesList()
	case "resources/read":
		resp.Result = handleResourcesRead(req.Params)
	case "prompts/list":
		resp.Result = handlePromptsList()
	case "prompts/get":
		resp.Result = handlePromptsGet(req.Params)
	default:
		resp.Error = map[string]interface{}{
			"code":    -32601,
			"message": "Method not found",
		}
	}

	json.NewEncoder(w).Encode(resp)
}

func handleInitialize(params interface{}) interface{} {
	return map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"capabilities": map[string]interface{}{
			"tools":     map[string]interface{}{},
			"resources": map[string]interface{}{},
			"prompts":   map[string]interface{}{},
		},
		"serverInfo": map[string]interface{}{
			"name":    "stepwise-test-server",
			"version": "1.0.0",
		},
	}
}

func handleToolsList() interface{} {
	return map[string]interface{}{
		"tools": []Tool{
			{
				Name:        "weather",
				Description: "Get weather information for a location",
				InputSchema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"location": map[string]interface{}{
							"type":        "string",
							"description": "City name or location",
						},
						"units": map[string]interface{}{
							"type":        "string",
							"enum":        []string{"celsius", "fahrenheit"},
							"description": "Temperature units",
						},
					},
					"required": []string{"location"},
				},
			},
			{
				Name:        "calculator",
				Description: "Perform mathematical calculations",
				InputSchema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"expression": map[string]interface{}{
							"type":        "string",
							"description": "Mathematical expression to evaluate",
						},
					},
					"required": []string{"expression"},
				},
			},
		},
	}
}

func handleToolsCall(params interface{}) interface{} {
	paramsMap, ok := params.(map[string]interface{})
	if !ok {
		return map[string]interface{}{
			"content": []Content{
				{
					Type: "text",
					Text: "Error: invalid parameters",
				},
			},
			"isError": true,
		}
	}

	toolName, _ := paramsMap["name"].(string)
	arguments, _ := paramsMap["arguments"].(map[string]interface{})

	switch toolName {
	case "weather":
		location, _ := arguments["location"].(string)
		units, _ := arguments["units"].(string)
		if units == "" {
			units = "celsius"
		}

		// Mock weather data
		weatherData := map[string]interface{}{
			"location":    location,
			"temperature": 22,
			"units":       units,
			"condition":   "sunny",
			"humidity":    65,
		}

		weatherJSON, _ := json.MarshalIndent(weatherData, "", "  ")

		return map[string]interface{}{
			"content": []Content{
				{
					Type: "text",
					Text: fmt.Sprintf("Weather in %s: %d°%s, %s, humidity %d%%\n\nJSON:\n%s",
						location, weatherData["temperature"],
						units[:1], weatherData["condition"],
						weatherData["humidity"], string(weatherJSON)),
				},
			},
			"isError": false,
		}

	case "calculator":
		expression, _ := arguments["expression"].(string)
		// Simple mock calculation
		result := fmt.Sprintf("Result of '%s' = 42 (mock result)", expression)

		return map[string]interface{}{
			"content": []Content{
				{
					Type: "text",
					Text: result,
				},
			},
			"isError": false,
		}

	default:
		return map[string]interface{}{
			"content": []Content{
				{
					Type: "text",
					Text: fmt.Sprintf("Unknown tool: %s", toolName),
				},
			},
			"isError": true,
		}
	}
}

func handleResourcesList() interface{} {
	return map[string]interface{}{
		"resources": []Resource{
			{
				URI:         "file:///tmp/data.json",
				Name:        "Test Data",
				Description: "Sample JSON data file",
				MimeType:    "application/json",
			},
			{
				URI:         "file:///tmp/config.yaml",
				Name:        "Configuration",
				Description: "Configuration file",
				MimeType:    "text/yaml",
			},
		},
	}
}

func handleResourcesRead(params interface{}) interface{} {
	paramsMap, ok := params.(map[string]interface{})
	if !ok {
		return map[string]interface{}{
			"contents": []Content{
				{
					Type: "text",
					Text: "Error: invalid parameters",
				},
			},
		}
	}

	uri, _ := paramsMap["uri"].(string)

	// Mock resource content based on URI
	var content string
	switch uri {
	case "file:///tmp/data.json":
		content = `{
  "name": "Test Data",
  "value": 42,
  "items": ["item1", "item2", "item3"]
}`
	case "file:///tmp/config.yaml":
		content = `name: Test Config
version: 1.0.0
settings:
  enabled: true
  timeout: 30s`
	default:
		content = fmt.Sprintf("Mock content for resource: %s\nTimestamp: %s", uri, time.Now().Format(time.RFC3339))
	}

	return map[string]interface{}{
		"contents": []Content{
			{
				Type: "text",
				Text: content,
			},
		},
	}
}

func handlePromptsList() interface{} {
	return map[string]interface{}{
		"prompts": []map[string]interface{}{
			{
				"name":        "code_review",
				"description": "Review code and provide feedback",
				"arguments": []map[string]interface{}{
					{
						"name":        "code",
						"description": "Code to review",
						"required":    true,
					},
					{
						"name":        "language",
						"description": "Programming language",
						"required":    false,
					},
				},
			},
			{
				"name":        "explain_code",
				"description": "Explain what code does",
				"arguments": []map[string]interface{}{
					{
						"name":        "code",
						"description": "Code to explain",
						"required":    true,
					},
				},
			},
		},
	}
}

func handlePromptsGet(params interface{}) interface{} {
	paramsMap, ok := params.(map[string]interface{})
	if !ok {
		return map[string]interface{}{
			"description": "Error: invalid parameters",
			"messages":    []interface{}{},
		}
	}

	promptName, _ := paramsMap["name"].(string)
	arguments, _ := paramsMap["arguments"].(map[string]interface{})

	switch promptName {
	case "code_review":
		code, _ := arguments["code"].(string)
		language, _ := arguments["language"].(string)
		if language == "" {
			language = "unknown"
		}

		return map[string]interface{}{
			"description": "Code review prompt",
			"messages": []map[string]interface{}{
				{
					"role": "user",
					"content": fmt.Sprintf("Please review the following %s code:\n\n```%s\n%s\n```\n\nProvide feedback on code quality, potential bugs, and improvements.",
						language, language, code),
				},
			},
		}

	case "explain_code":
		code, _ := arguments["code"].(string)

		return map[string]interface{}{
			"description": "Code explanation prompt",
			"messages": []map[string]interface{}{
				{
					"role":    "user",
					"content": fmt.Sprintf("Please explain what the following code does:\n\n```\n%s\n```", code),
				},
			},
		}

	default:
		return map[string]interface{}{
			"description": fmt.Sprintf("Unknown prompt: %s", promptName),
			"messages":    []interface{}{},
		}
	}
}

func respondError(w http.ResponseWriter, id interface{}, code int, message, data string) {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: map[string]interface{}{
			"code":    code,
			"message": message,
			"data":    data,
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
