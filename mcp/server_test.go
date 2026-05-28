package mcpserver_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	mcpserver "github.com/pouyasadri/env-validator-mcp/mcp"
)

// buildTestServer creates an MCP server wired with mock dependencies.
func buildTestServer(t *testing.T) (*mcp.Server, *mcp.ClientSession) {
	t.Helper()

	deps := mcpserver.TestDependencies()
	server := mcpserver.NewServer("test", deps)

	ctx := context.Background()
	t1, t2 := mcp.NewInMemoryTransports()

	_, err := server.Connect(ctx, t1, nil)
	require.NoError(t, err)

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "v1.0.0"}, nil)
	session, err := client.Connect(ctx, t2, nil)
	require.NoError(t, err)

	t.Cleanup(func() { session.Close() })
	return server, session
}

func TestListTools(t *testing.T) {
	_, session := buildTestServer(t)

	ctx := context.Background()
	result, err := session.ListTools(ctx, nil)
	require.NoError(t, err)

	names := make([]string, 0, len(result.Tools))
	for _, tool := range result.Tools {
		names = append(names, tool.Name)
	}

	assert.Contains(t, names, "check_tool_version")
	assert.Contains(t, names, "check_env_vars")
	assert.Contains(t, names, "check_config_files")
	assert.Contains(t, names, "validate_environment")
	assert.Contains(t, names, "list_checks")
}

func TestTool_CheckToolVersion_OK(t *testing.T) {
	_, session := buildTestServer(t)
	ctx := context.Background()

	res, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "check_tool_version",
		Arguments: map[string]any{
			"tool":       "go",
			"constraint": ">=1.21.0",
		},
	})
	require.NoError(t, err)
	require.False(t, res.IsError, "tool call should not be an error")
	require.NotEmpty(t, res.Content)

	text := res.Content[0].(*mcp.TextContent).Text
	assert.Contains(t, text, `"status"`)
	assert.Contains(t, text, `"ok"`)
}

func TestTool_CheckToolVersion_Missing(t *testing.T) {
	_, session := buildTestServer(t)
	ctx := context.Background()

	res, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "check_tool_version",
		Arguments: map[string]any{"tool": "docker"},
	})
	require.NoError(t, err)
	require.False(t, res.IsError)

	text := res.Content[0].(*mcp.TextContent).Text
	assert.Contains(t, text, `"missing"`)
}

func TestTool_CheckToolVersion_UnknownTool(t *testing.T) {
	_, session := buildTestServer(t)
	ctx := context.Background()

	res, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "check_tool_version",
		Arguments: map[string]any{"tool": "does-not-exist"},
	})
	require.NoError(t, err)
	assert.True(t, res.IsError, "unknown tool should return an MCP error result")
}

func TestTool_CheckEnvVars_OK(t *testing.T) {
	_, session := buildTestServer(t)
	ctx := context.Background()

	res, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "check_env_vars",
		Arguments: map[string]any{
			"vars": []any{"HOME"},
		},
	})
	require.NoError(t, err)
	require.False(t, res.IsError)

	text := res.Content[0].(*mcp.TextContent).Text
	assert.Contains(t, text, "HOME")
	assert.Contains(t, text, `"ok"`)
}

func TestTool_CheckEnvVars_Missing(t *testing.T) {
	_, session := buildTestServer(t)
	ctx := context.Background()

	res, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "check_env_vars",
		Arguments: map[string]any{
			"vars": []any{"DEFINITELY_NOT_SET_XYZ_123"},
		},
	})
	require.NoError(t, err)
	require.False(t, res.IsError)

	text := res.Content[0].(*mcp.TextContent).Text
	assert.Contains(t, text, `"missing"`)
}

func TestTool_CheckConfigFiles_OK(t *testing.T) {
	_, session := buildTestServer(t)
	ctx := context.Background()

	// Use /tmp which always exists
	res, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "check_config_files",
		Arguments: map[string]any{
			"paths": []any{"/tmp"},
		},
	})
	require.NoError(t, err)
	require.False(t, res.IsError)

	text := res.Content[0].(*mcp.TextContent).Text
	assert.Contains(t, text, `"ok"`)
}

func TestTool_CheckConfigFiles_Missing(t *testing.T) {
	_, session := buildTestServer(t)
	ctx := context.Background()

	res, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "check_config_files",
		Arguments: map[string]any{
			"paths": []any{"/this/path/does/not/exist/abc123"},
		},
	})
	require.NoError(t, err)
	require.False(t, res.IsError)

	text := res.Content[0].(*mcp.TextContent).Text
	assert.Contains(t, text, `"missing"`)
}

func TestTool_ValidateEnvironment(t *testing.T) {
	_, session := buildTestServer(t)
	ctx := context.Background()

	res, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "validate_environment",
		Arguments: map[string]any{},
	})
	require.NoError(t, err)
	require.False(t, res.IsError)

	text := res.Content[0].(*mcp.TextContent).Text

	// Must be valid JSON with a summary field
	var report map[string]any
	require.NoError(t, json.Unmarshal([]byte(text), &report), "response must be valid JSON")
	assert.Contains(t, report, "summary")
	assert.Contains(t, report, "results")
	assert.Contains(t, report, "ok_count")
	assert.Contains(t, report, "issue_count")
}

func TestTool_ListChecks(t *testing.T) {
	_, session := buildTestServer(t)
	ctx := context.Background()

	res, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "list_checks",
		Arguments: map[string]any{},
	})
	require.NoError(t, err)
	require.False(t, res.IsError)

	text := res.Content[0].(*mcp.TextContent).Text
	assert.Contains(t, text, "go")
	assert.Contains(t, text, "docker")
	assert.Contains(t, text, "node")
}
