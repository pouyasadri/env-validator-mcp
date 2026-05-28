// Command server is the env-validator MCP server.
// It exposes tools that inspect the local development environment and report
// missing dependencies, version mismatches, and misconfigured environment variables.
//
// Usage (stdio transport — compatible with Claude Desktop, Cursor, Gemini CLI):
//
//	./env-validator-mcp
//
// Configure in Claude Desktop (~/.config/claude/claude_desktop_config.json):
//
//	{
//	  "mcpServers": {
//	    "env-validator": {
//	      "command": "/path/to/env-validator-mcp"
//	    }
//	  }
//	}
package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	mcpserver "github.com/pouyasadri/env-validator-mcp/mcp"
)

const version = "1.0.0"

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	logger.Info("starting env-validator MCP server", "version", version)

	deps := mcpserver.DefaultDependencies()
	server := mcpserver.NewServer(version, deps)

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		logger.Error("server exited with error", "err", err)
		os.Exit(1)
	}
}
