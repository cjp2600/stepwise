package main

import (
	"context"
	"fmt"
	"github.com/cjp2600/stepwise/internal/ai"
)

func main() {
	// Create client
	client := ai.NewClient()
	client.SetWorkingDir("./examples")

	// Build context
	contextPrompt, err := ai.BuildContextPrompt("./examples")
	if err != nil {
		fmt.Printf("Error building context: %v\n", err)
		return
	}

	// Show first 1000 characters of system prompt
	if len(contextPrompt) > 1000 {
		fmt.Printf("System prompt (first 1000 chars):\n%s...\n\n", contextPrompt[:1000])
	} else {
		fmt.Printf("System prompt:\n%s\n\n", contextPrompt)
	}

	// Create messages
	messages := []ai.Message{
		{Role: "system", Content: contextPrompt},
		{Role: "user", Content: "Создай простой workflow для тестирования API пользователей"},
	}

	// Try to get response
	fmt.Println("Sending request to codex...")
	ctx := context.Background()
	
	err = client.ChatStream(ctx, messages, func(chunk string) {
		fmt.Print(chunk)
	})
	
	if err != nil {
		fmt.Printf("\nError: %v\n", err)
	}
}
