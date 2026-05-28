package mcpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/pouyasadri/env-validator-mcp/internal/checker"
	"github.com/pouyasadri/env-validator-mcp/internal/report"
)

// handlers holds all handler functions and their shared dependencies.
type handlers struct {
	toolChecker   *checker.ToolChecker
	envChecker    *checker.EnvChecker
	configChecker *checker.ConfigChecker
}

func newHandlers(deps Dependencies) *handlers {
	return &handlers{
		toolChecker:   checker.NewToolChecker(deps.Commander),
		envChecker:    checker.NewEnvChecker(deps.EnvReader),
		configChecker: checker.NewConfigChecker(deps.FileSystem),
	}
}

// --- Input types (auto-generate JSON Schema via mcp.AddTool generic) ---

type checkToolVersionInput struct {
	Tool       string `json:"tool"`
	Constraint string `json:"constraint,omitempty"`
}

type checkEnvVarsInput struct {
	Vars            []string `json:"vars"`
	RequireNonEmpty bool     `json:"require_non_empty,omitempty"`
	Pattern         string   `json:"pattern,omitempty"`
}

type checkConfigFilesInput struct {
	Paths        []string `json:"paths"`
	ValidateJSON bool     `json:"validate_json,omitempty"`
	RequiredKeys []string `json:"required_keys,omitempty"`
}

type validateEnvironmentInput struct{}

type listChecksInput struct{}

// --- Handlers ---

// checkToolVersion checks whether a single CLI tool is installed and meets a version constraint.
func (h *handlers) checkToolVersion(
	ctx context.Context,
	_ *mcp.CallToolRequest,
	in checkToolVersionInput,
) (*mcp.CallToolResult, any, error) {
	if in.Tool == "" {
		return errorResult("tool name is required"), nil, nil
	}

	result, err := h.toolChecker.Check(ctx, in.Tool, in.Constraint)
	if err != nil {
		return errorResult(err.Error()), nil, nil
	}

	return jsonResult(result)
}

// checkEnvVars checks a list of environment variables.
func (h *handlers) checkEnvVars(
	_ context.Context,
	_ *mcp.CallToolRequest,
	in checkEnvVarsInput,
) (*mcp.CallToolResult, any, error) {
	if len(in.Vars) == 0 {
		return errorResult("at least one variable name is required in 'vars'"), nil, nil
	}

	spec := checker.EnvSpec{
		RequireNonEmpty: in.RequireNonEmpty,
		Pattern:         in.Pattern,
	}

	specs := make(map[string]checker.EnvSpec, len(in.Vars))
	for _, v := range in.Vars {
		specs[v] = spec
	}

	results := h.envChecker.CheckMultiple(specs)
	rep := report.NewReport(results)
	return jsonResult(rep)
}

// checkConfigFiles checks whether specified file paths exist.
func (h *handlers) checkConfigFiles(
	_ context.Context,
	_ *mcp.CallToolRequest,
	in checkConfigFilesInput,
) (*mcp.CallToolResult, any, error) {
	if len(in.Paths) == 0 {
		return errorResult("at least one path is required in 'paths'"), nil, nil
	}

	spec := checker.FileSpec{
		ValidateJSON: in.ValidateJSON,
		RequiredKeys: in.RequiredKeys,
	}

	specs := make(map[string]checker.FileSpec, len(in.Paths))
	for _, p := range in.Paths {
		specs[p] = spec
	}

	results := h.configChecker.CheckMultiple(specs)
	rep := report.NewReport(results)
	return jsonResult(rep)
}

// validateEnvironment runs a comprehensive sweep of the developer environment.
func (h *handlers) validateEnvironment(
	ctx context.Context,
	_ *mcp.CallToolRequest,
	_ validateEnvironmentInput,
) (*mcp.CallToolResult, any, error) {
	var allResults []report.Result

	// --- Tool checks ---
	toolSpecs := []checker.ToolSpec{
		{Name: "go", Constraint: ">=1.21.0"},
		{Name: "node", Constraint: ""},
		{Name: "npm", Constraint: ""},
		{Name: "docker", Constraint: ""},
		{Name: "git", Constraint: ">=2.30.0"},
		{Name: "python3", Constraint: ""},
		{Name: "make", Constraint: ""},
		{Name: "curl", Constraint: ""},
	}
	toolResults, err := h.toolChecker.CheckMultiple(ctx, toolSpecs)
	if err != nil {
		return errorResult(fmt.Sprintf("tool check failed: %v", err)), nil, nil
	}
	allResults = append(allResults, toolResults...)

	// --- Environment variable checks ---
	envSpecs := map[string]checker.EnvSpec{
		"HOME":   {RequireNonEmpty: true},
		"PATH":   {RequireNonEmpty: true},
		"GOPATH": {},
		"GOROOT": {},
	}
	allResults = append(allResults, h.envChecker.CheckMultiple(envSpecs)...)

	// --- Config file checks ---
	configSpecs := map[string]checker.FileSpec{
		"go.mod":             {},
		"package.json":       {ValidateJSON: true},
		".env":               {},
		"docker-compose.yml": {},
		"Dockerfile":         {},
		".gitignore":         {},
	}
	allResults = append(allResults, h.configChecker.CheckMultiple(configSpecs)...)

	rep := report.NewReport(allResults)
	return jsonResult(rep)
}

// listChecks returns the catalog of everything this server can check.
func (h *handlers) listChecks(
	_ context.Context,
	_ *mcp.CallToolRequest,
	_ listChecksInput,
) (*mcp.CallToolResult, any, error) {
	tools := checker.SupportedTools()
	sort.Strings(tools)

	catalog := map[string]any{
		"tools": tools,
		"env_vars": []string{
			"HOME", "PATH", "GOPATH", "GOROOT", "GOBIN",
			"DATABASE_URL", "API_KEY", "PORT", "NODE_ENV",
			"AWS_REGION", "AWS_ACCESS_KEY_ID", "KUBECONFIG",
		},
		"config_files": []string{
			"go.mod", "go.sum", "package.json", "package-lock.json",
			".env", ".env.local", ".env.production",
			"docker-compose.yml", "Dockerfile",
			".gitignore", ".github/workflows/",
			"Makefile", "terraform.tfvars", "helm/Chart.yaml",
		},
		"constraints_syntax": "Supported operators: >=, >, <=, <, = (e.g. '>=1.21.0', '<3.0.0')",
	}

	return jsonResult(catalog)
}

// --- Helpers ---

// jsonResult marshals v to pretty JSON and wraps it in a TextContent result.
func jsonResult(v any) (*mcp.CallToolResult, any, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return errorResult(fmt.Sprintf("failed to marshal result: %v", err)), nil, nil
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(data)},
		},
	}, nil, nil
}

// errorResult returns an MCP error result with the given message.
func errorResult(message string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{
			&mcp.TextContent{Text: message},
		},
	}
}
