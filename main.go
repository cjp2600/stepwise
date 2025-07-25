package main

import (
	"fmt"
	"os"

	"github.com/cjp2600/stepwise/internal/cli"
	"github.com/cjp2600/stepwise/internal/config"
	"github.com/cjp2600/stepwise/internal/logger"
)

func main() {
	// Initialize logger
	logger := logger.New()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Create CLI application
	app := cli.NewApp(cfg, logger)

	// Run the CLI
	if err := app.Run(os.Args); err != nil {
		fmt.Printf("\n✗ %v\n", err)
		os.Exit(1)
	}

	// If we reach here, everything was successful
	os.Exit(0)
}
