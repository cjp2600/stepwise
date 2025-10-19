package ai

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ChatInterface represents an interactive chat interface
type ChatInterface struct {
	client    *Client
	messages  []Message
	colors    ChatColors
	outputDir string
}

// ChatColors provides color output for chat
type ChatColors struct {
	UserPrompt      string
	AssistantPrompt string
	UserMessage     string
	AssistantMsg    string
	SystemMessage   string
	ErrorMessage    string
	Reset           string
}

// NewChatInterface creates a new chat interface
func NewChatInterface(client *Client) *ChatInterface {
	return &ChatInterface{
		client:    client,
		messages:  make([]Message, 0),
		outputDir: ".",
		colors: ChatColors{
			UserPrompt:      "\033[1;36m", // Cyan bold
			AssistantPrompt: "\033[1;32m", // Green bold
			UserMessage:     "\033[0;36m", // Cyan
			AssistantMsg:    "\033[0;32m", // Green
			SystemMessage:   "\033[0;33m", // Yellow
			ErrorMessage:    "\033[0;31m", // Red
			Reset:           "\033[0m",    // Reset
		},
	}
}

// SetOutputDirectory sets the directory where workflows will be saved
func (c *ChatInterface) SetOutputDirectory(dir string) {
	c.outputDir = dir
}

// AddSystemMessage adds a system message to the conversation
func (c *ChatInterface) AddSystemMessage(content string) {
	c.messages = append(c.messages, Message{
		Role:    "system",
		Content: content,
	})
}

// AddUserMessage adds a user message to the conversation
func (c *ChatInterface) AddUserMessage(content string) {
	c.messages = append(c.messages, Message{
		Role:    "user",
		Content: content,
	})
}

// AddAssistantMessage adds an assistant message to the conversation
func (c *ChatInterface) AddAssistantMessage(content string) {
	c.messages = append(c.messages, Message{
		Role:    "assistant",
		Content: content,
	})
}

// Start starts the interactive chat session
func (c *ChatInterface) Start(ctx context.Context) error {
	fmt.Printf("%s╔══════════════════════════════════════════════════════════════════╗%s\n", c.colors.AssistantPrompt, c.colors.Reset)
	fmt.Printf("%s║%s      Stepwise AI Assistant - Powered by OpenAI                   %s║%s\n", c.colors.AssistantPrompt, c.colors.Reset, c.colors.AssistantPrompt, c.colors.Reset)
	fmt.Printf("%s╚══════════════════════════════════════════════════════════════════╝%s\n\n", c.colors.AssistantPrompt, c.colors.Reset)

	fmt.Printf("%sI'm your AI assistant for creating and managing Stepwise workflows.%s\n", c.colors.SystemMessage, c.colors.Reset)
	fmt.Printf("%sI have analyzed your existing workflows and components.%s\n", c.colors.SystemMessage, c.colors.Reset)
	fmt.Printf("%sI can create files directly using codex CLI or use /save for manual saving.%s\n", c.colors.SystemMessage, c.colors.Reset)
	fmt.Printf("%sI will ask codex to analyze all files in this directory for better context.%s\n\n", c.colors.SystemMessage, c.colors.Reset)

	fmt.Printf("%sCommands:%s\n", c.colors.SystemMessage, c.colors.Reset)
	fmt.Printf("  %s/help%s     - Show this help message\n", c.colors.UserMessage, c.colors.Reset)
	fmt.Printf("  %s/clear%s    - Clear conversation history\n", c.colors.UserMessage, c.colors.Reset)
	fmt.Printf("  %s/save%s     - Save last workflow to file\n", c.colors.UserMessage, c.colors.Reset)
	fmt.Printf("  %s/exit%s     - Exit chat\n", c.colors.UserMessage, c.colors.Reset)
	fmt.Printf("  %s/quit%s     - Exit chat\n\n", c.colors.UserMessage, c.colors.Reset)

	fmt.Printf("%sWhat would you like to do?%s\n", c.colors.SystemMessage, c.colors.Reset)
	fmt.Printf("%s- Create a new workflow%s\n", c.colors.SystemMessage, c.colors.Reset)
	fmt.Printf("%s- Improve an existing workflow%s\n", c.colors.SystemMessage, c.colors.Reset)
	fmt.Printf("%s- Create a component%s\n", c.colors.SystemMessage, c.colors.Reset)
	fmt.Printf("%s- Ask questions about your test suite%s\n\n", c.colors.SystemMessage, c.colors.Reset)

	scanner := bufio.NewScanner(os.Stdin)
	lastWorkflow := ""

	for {
		// User prompt
		fmt.Printf("%s╭─[You]%s\n", c.colors.UserPrompt, c.colors.Reset)
		fmt.Printf("%s╰─➤%s ", c.colors.UserPrompt, c.colors.Reset)

		if !scanner.Scan() {
			break
		}

		userInput := strings.TrimSpace(scanner.Text())
		if userInput == "" {
			continue
		}

		// Handle commands
		if strings.HasPrefix(userInput, "/") {
			switch strings.ToLower(userInput) {
			case "/exit", "/quit":
				fmt.Printf("\n%s👋 Goodbye! Happy testing!%s\n", c.colors.SystemMessage, c.colors.Reset)
				return nil
			case "/clear":
				// Keep only system message
				if len(c.messages) > 0 && c.messages[0].Role == "system" {
					c.messages = c.messages[:1]
				} else {
					c.messages = make([]Message, 0)
				}
				fmt.Printf("\n%s✓ Conversation history cleared%s\n\n", c.colors.SystemMessage, c.colors.Reset)
				continue
			case "/help":
				c.showHelp()
				continue
			case "/save":
				if lastWorkflow != "" {
					if err := c.saveWorkflow(lastWorkflow); err != nil {
						fmt.Printf("\n%s✗ Error saving workflow: %v%s\n\n", c.colors.ErrorMessage, err, c.colors.Reset)
					}
				} else {
					fmt.Printf("\n%s✗ No workflow to save%s\n\n", c.colors.ErrorMessage, c.colors.Reset)
				}
				continue
			default:
				fmt.Printf("\n%s✗ Unknown command. Type /help for available commands%s\n\n", c.colors.ErrorMessage, c.colors.Reset)
				continue
			}
		}

		// Add user message
		c.AddUserMessage(userInput)

		// Show thinking indicator
		fmt.Printf("\n%s╭─[Assistant]%s\n", c.colors.AssistantPrompt, c.colors.Reset)
		fmt.Printf("%s╰─➤%s %sThinking...%s", c.colors.AssistantPrompt, c.colors.Reset, c.colors.SystemMessage, c.colors.Reset)

		// Stream response
		var responseBuilder strings.Builder
		firstChunk := true
		err := c.client.ChatStream(ctx, c.messages, func(chunk string) {
			if firstChunk {
				// Clear "Thinking..." message
				fmt.Printf("\r%s╰─➤%s ", c.colors.AssistantPrompt, c.colors.Reset)
				firstChunk = false
			}
			fmt.Print(chunk)
			responseBuilder.WriteString(chunk)
		})

		if err != nil {
			fmt.Printf("\r%s╰─➤%s %s✗ Error: %v%s\n\n", c.colors.AssistantPrompt, c.colors.Reset, c.colors.ErrorMessage, err, c.colors.Reset)
			// Remove last user message since we failed
			if len(c.messages) > 0 {
				c.messages = c.messages[:len(c.messages)-1]
			}
			continue
		}

		response := responseBuilder.String()

		// Check if we got any response
		if response == "" {
			fmt.Printf("\r%s╰─➤%s %s✗ No response received from AI%s\n\n", c.colors.AssistantPrompt, c.colors.Reset, c.colors.ErrorMessage, c.colors.Reset)
			// Remove last user message since we got no response
			if len(c.messages) > 0 {
				c.messages = c.messages[:len(c.messages)-1]
			}
			continue
		}

		fmt.Printf("\n\n")

		// Add assistant response to history
		c.AddAssistantMessage(response)

		// Extract workflow if present (for manual save only)
		if workflow := c.extractYAML(response); workflow != "" {
			lastWorkflow = workflow
			fmt.Printf("\n%s💡 Workflow detected! Use /save to save it to a file.%s\n\n", c.colors.SystemMessage, c.colors.Reset)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading input: %w", err)
	}

	return nil
}

// showHelp displays help information
func (c *ChatInterface) showHelp() {
	fmt.Printf("\n%s╔══════════════════════════════════════════════════════════════════╗%s\n", c.colors.SystemMessage, c.colors.Reset)
	fmt.Printf("%s║%s                       HELP - Commands                       %s║%s\n", c.colors.SystemMessage, c.colors.Reset, c.colors.SystemMessage, c.colors.Reset)
	fmt.Printf("%s╚══════════════════════════════════════════════════════════════════╝%s\n\n", c.colors.SystemMessage, c.colors.Reset)

	fmt.Printf("%sAvailable Commands:%s\n", c.colors.SystemMessage, c.colors.Reset)
	fmt.Printf("  %s/help%s     - Show this help message\n", c.colors.UserMessage, c.colors.Reset)
	fmt.Printf("  %s/clear%s    - Clear conversation history (keeps system context)\n", c.colors.UserMessage, c.colors.Reset)
	fmt.Printf("  %s/save%s     - Save the last generated workflow to a file\n", c.colors.UserMessage, c.colors.Reset)
	fmt.Printf("  %s/exit%s     - Exit the chat interface\n", c.colors.UserMessage, c.colors.Reset)
	fmt.Printf("  %s/quit%s     - Exit the chat interface\n\n", c.colors.UserMessage, c.colors.Reset)

	fmt.Printf("%sExample Prompts:%s\n", c.colors.SystemMessage, c.colors.Reset)
	fmt.Printf("  • Create a workflow to test user registration API\n")
	fmt.Printf("  • Add validation to check if email format is correct\n")
	fmt.Printf("  • Create a component for authentication\n")
	fmt.Printf("  • How do I capture nested JSON fields?\n")
	fmt.Printf("  • Improve the performance-test.yml workflow\n\n")
}

// extractYAML extracts YAML content from markdown code blocks
func (c *ChatInterface) extractYAML(text string) string {
	lines := strings.Split(text, "\n")
	var yamlContent []string
	inYAML := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```yaml") || strings.HasPrefix(trimmed, "```yml") {
			inYAML = true
			continue
		}
		if inYAML && strings.HasPrefix(trimmed, "```") {
			break
		}
		if inYAML {
			yamlContent = append(yamlContent, line)
		}
	}

	return strings.Join(yamlContent, "\n")
}

// autoSaveWorkflow automatically saves a workflow to a file with a generated name
func (c *ChatInterface) autoSaveWorkflow(content string) (string, error) {
	// Extract workflow name from YAML content
	name := c.extractWorkflowName(content)
	if name == "" {
		name = "generated-workflow"
	}

	// Generate filename based on workflow name
	filename := c.generateFilename(name)

	// Ensure unique filename
	filename = c.ensureUniqueFilename(filename)

	// Create full path
	fullPath := filepath.Join(c.outputDir, filename)

	err := os.WriteFile(fullPath, []byte(content), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return fullPath, nil
}

// extractWorkflowName extracts the name from YAML content
func (c *ChatInterface) extractWorkflowName(content string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "name:") {
			// Extract name value
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				name := strings.Trim(strings.TrimSpace(parts[1]), "\"'")
				return name
			}
		}
	}
	return ""
}

// generateFilename generates a filename from workflow name
func (c *ChatInterface) generateFilename(name string) string {
	// Convert to lowercase and replace spaces/special chars with hyphens
	filename := strings.ToLower(name)
	filename = strings.ReplaceAll(filename, " ", "-")
	filename = strings.ReplaceAll(filename, "_", "-")

	// Remove special characters
	var result strings.Builder
	for _, char := range filename {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '-' {
			result.WriteRune(char)
		}
	}

	filename = result.String()

	// Ensure it's not empty
	if filename == "" {
		filename = "workflow"
	}

	return filename + ".yml"
}

// ensureUniqueFilename ensures the filename is unique by adding a number if needed
func (c *ChatInterface) ensureUniqueFilename(filename string) string {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return filename
	}

	// File exists, try with numbers
	base := strings.TrimSuffix(filename, ".yml")
	for i := 1; i <= 100; i++ {
		newFilename := fmt.Sprintf("%s-%d.yml", base, i)
		if _, err := os.Stat(newFilename); os.IsNotExist(err) {
			return newFilename
		}
	}

	// Fallback to timestamp
	timestamp := time.Now().Format("20060102-150405")
	return fmt.Sprintf("%s-%s.yml", base, timestamp)
}

// saveWorkflow saves a workflow to a file
func (c *ChatInterface) saveWorkflow(content string) error {
	fmt.Printf("\n%sEnter filename (e.g., my-workflow.yml): %s", c.colors.SystemMessage, c.colors.Reset)

	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return fmt.Errorf("failed to read filename")
	}

	filename := strings.TrimSpace(scanner.Text())
	if filename == "" {
		return fmt.Errorf("filename cannot be empty")
	}

	// Ensure .yml extension
	if !strings.HasSuffix(filename, ".yml") && !strings.HasSuffix(filename, ".yaml") {
		filename += ".yml"
	}

	err := os.WriteFile(filename, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	fmt.Printf("\n%s✓ Workflow saved to %s%s\n\n", c.colors.SystemMessage, filename, c.colors.Reset)
	return nil
}
