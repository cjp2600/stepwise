package main

import (
	"context"
	"os"

	"github.com/cjp2600/stepwise/internal/cli"
	"github.com/cjp2600/stepwise/internal/config"
	"github.com/cjp2600/stepwise/internal/logger"
	"github.com/cjp2600/stepwise/internal/mcp"
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

	// Check if MCP server mode is requested
	if len(os.Args) > 1 && (os.Args[1] == "--mcp-server" || os.Args[1] == "-mcp") {
		// Run as MCP server
		cliApp := cli.NewApp(cfg, logger)
		server := mcp.NewServer(cfg, logger, cliApp)
		ctx := context.Background()
		if err := server.Run(ctx); err != nil {
			logger.Error("MCP server failed", "error", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	// Create CLI application
	app := cli.NewApp(cfg, logger)

	// Run the CLI
	if err := app.Run(os.Args); err != nil {
		logger.Error("CLI execution failed", "error", err)
		os.Exit(1)
	}

	// If we reach here, everything was successful
	os.Exit(0)
}
