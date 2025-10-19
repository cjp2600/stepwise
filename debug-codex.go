package main

import (
	"context"
	"fmt"
	"os"

	"github.com/cjp2600/stepwise/internal/ai"
)

func main() {
	client := ai.NewClient()
	client.SetModel("o1")
	client.SetWorkingDir(".")

	// Simple test without system prompt
	messages := []ai.Message{
		{
			Role:    "user",
			Content: "Привет! Как дела?",
		},
	}

	fmt.Println("Testing codex with simple prompt...")

	// Test streaming
	err := client.ChatStream(context.Background(), messages, func(chunk string) {
		fmt.Printf("Chunk: %s\n", chunk)
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
