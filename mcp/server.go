// Package mcpserver builds and configures the MCP server with all tools.
package mcpserver

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/pouyasadri/env-validator-mcp/internal/checker"
)

// Dependencies bundles all injectable dependencies for the MCP server.
type Dependencies struct {
	Commander  checker.Commander
	EnvReader  checker.EnvReader
	FileSystem checker.FileSystem
}

// DefaultDependencies returns production implementations that call the real OS.
func DefaultDependencies() Dependencies {
	return Dependencies{
		Commander:  checker.NewRealCommander(),
		EnvReader:  checker.NewRealEnvReader(),
		FileSystem: checker.NewRealFileSystem(),
	}
}

// NewServer constructs and returns a fully configured MCP server.
// All tools are registered and annotated as read-only (they never mutate the system).
func NewServer(version string, deps Dependencies) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "env-validator",
		Version: version,
	}, nil)

	h := newHandlers(deps)

	annotations := &mcp.ToolAnnotations{ReadOnlyHint: true}

	// check_tool_version — verify a CLI tool is installed and optionally meets a semver constraint.
	mcp.AddTool(server, &mcp.Tool{
		Name:        "check_tool_version",
		Description: "Check whether a CLI tool is installed and optionally satisfies a semver version constraint (e.g. '>=1.21.0'). Supported tools: go, node, npm, docker, git, python3, pip3, make, curl, jq, kubectl, terraform, helm, rustc, java.",
		Annotations: annotations,
	}, h.checkToolVersion)

	// check_env_vars — verify environment variables are set and optionally match a pattern.
	mcp.AddTool(server, &mcp.Tool{
		Name:        "check_env_vars",
		Description: "Check whether one or more environment variables are set. Optionally require non-empty values or validate values against a regular expression pattern.",
		Annotations: annotations,
	}, h.checkEnvVars)

	// check_config_files — verify config files/directories exist.
	mcp.AddTool(server, &mcp.Tool{
		Name:        "check_config_files",
		Description: "Check whether expected configuration files or directories exist on the file system. Optionally validate JSON files for syntactic correctness and required top-level keys.",
		Annotations: annotations,
	}, h.checkConfigFiles)

	// validate_environment — run a full environment sweep.
	mcp.AddTool(server, &mcp.Tool{
		Name:        "validate_environment",
		Description: "Run a comprehensive environment validation: checks common dev tools (go, node, npm, docker, git), critical environment variables (HOME, PATH, GOPATH), and common config files (.env, go.mod, package.json). Returns a structured JSON report with status, versions, and a summary.",
		Annotations: annotations,
	}, h.validateEnvironment)

	// list_checks — enumerate everything the server can inspect.
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_checks",
		Description: "List all tools, environment variables, and config files that this server is able to check.",
		Annotations: annotations,
	}, h.listChecks)

	return server
}
