package ai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// Client represents a Codex CLI client
type Client struct {
	model      string
	workingDir string
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// CodexEvent represents an event from codex JSON output
type CodexEvent struct {
	Msg struct {
		Type    string `json:"type"`
		Message string `json:"message"`
		Content string `json:"content"`
		Text    string `json:"text"`
	} `json:"msg"`
}

// NewClient creates a new Codex CLI client
func NewClient() *Client {
	return &Client{
		model:      "gpt-5-codex", // Default model
		workingDir: ".",
	}
}

// SetModel sets the model to use
func (c *Client) SetModel(model string) {
	c.model = model
}

// SetWorkingDir sets the working directory for codex
func (c *Client) SetWorkingDir(dir string) {
	c.workingDir = dir
}

// Chat sends a chat completion request using codex exec
func (c *Client) Chat(ctx context.Context, messages []Message) (string, error) {
	// Build prompt from messages
	prompt := c.buildPrompt(messages)

	// Execute codex
	cmd := exec.CommandContext(ctx, "codex", "exec",
		"--model", c.model,
		"--json",
		"--skip-git-repo-check",
		"--output-last-message", "/dev/stdout",
		"--full-auto",
		"-C", c.workingDir,
		prompt,
	)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("codex exec failed: %w\nStderr: %s", err, stderr.String())
	}

	// Parse JSON output and extract last message
	return c.parseLastMessage(stdout.String())
}

// ChatStream sends a streaming chat completion request
func (c *Client) ChatStream(ctx context.Context, messages []Message, onChunk func(string)) error {
	// Build prompt from messages
	prompt := c.buildPrompt(messages)

	// Execute codex with JSON output
	cmd := exec.CommandContext(ctx, "codex", "exec",
		"--json",
		"--skip-git-repo-check",
		"--full-auto",
		"-C", c.workingDir,
		prompt,
	)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start codex: %w", err)
	}

	// Read stderr in background
	go func() {
		stderrScanner := bufio.NewScanner(stderr)
		for stderrScanner.Scan() {
			// Silently ignore stderr for now
			// fmt.Fprintf(os.Stderr, "CODEX STDERR: %s\n", stderrScanner.Text())
		}
	}()

	// Stream and parse JSON events
	scanner := bufio.NewScanner(stdout)

	// Increase buffer size to handle large JSON responses
	const maxCapacity = 10 * 1024 * 1024 // 10MB
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	for scanner.Scan() {
		line := scanner.Text()

		// Skip non-JSON lines
		if !strings.HasPrefix(line, "{") {
			continue
		}

		// Try to parse as JSON event
		var event CodexEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			continue
		}

		// Handle different types of agent events
		switch event.Msg.Type {
		case "agent_message":
			message := event.Msg.Message
			if message == "" {
				message = event.Msg.Content
			}
			if message == "" {
				message = event.Msg.Text
			}

			if message != "" {
				// Send the full response at once
				onChunk(message)
			}
		case "agent_reasoning":
			// Skip reasoning messages
			continue
		default:
			// Skip other message types
			continue
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading codex output: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("codex exec failed: %w", err)
	}

	return nil
}

// buildPrompt builds a complete prompt from message history
func (c *Client) buildPrompt(messages []Message) string {
	var parts []string

	for _, msg := range messages {
		switch msg.Role {
		case "system":
			parts = append(parts, fmt.Sprintf("# SYSTEM INSTRUCTIONS\n\n%s\n", msg.Content))
		case "user":
			parts = append(parts, fmt.Sprintf("# USER\n\n%s\n", msg.Content))
		case "assistant":
			parts = append(parts, fmt.Sprintf("# ASSISTANT\n\n%s\n", msg.Content))
		}
	}

	return strings.Join(parts, "\n")
}

// parseLastMessage extracts the last agent message from JSON output
func (c *Client) parseLastMessage(output string) (string, error) {
	lines := strings.Split(output, "\n")
	var lastMessage string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var event CodexEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			continue
		}

		if event.Msg.Type == "agent_message" {
			message := event.Msg.Message
			if message == "" {
				message = event.Msg.Content
			}
			if message != "" {
				lastMessage = message
			}
		}
	}

	if lastMessage == "" {
		return "", fmt.Errorf("no agent message found in output")
	}

	return lastMessage, nil
}
