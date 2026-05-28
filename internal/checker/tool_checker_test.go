package checker_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pouyasadri/env-validator-mcp/internal/checker"
	"github.com/pouyasadri/env-validator-mcp/internal/report"
)

func TestCheckTool_Installed_NoConstraint(t *testing.T) {
	cmd := NewMockCommander()
	cmd.Responses["go version"] = "go version go1.24.2 darwin/arm64"

	tc := checker.NewToolChecker(cmd)
	result, err := tc.Check(context.Background(), "go", "")

	require.NoError(t, err)
	assert.Equal(t, report.StatusOK, result.Status)
	assert.Equal(t, "go", result.Name)
	assert.Equal(t, "1.24.2", result.FoundVersion)
	assert.Empty(t, result.ExpectedRange)
}

func TestCheckTool_Installed_ConstraintSatisfied(t *testing.T) {
	cmd := NewMockCommander()
	cmd.Responses["node --version"] = "v20.11.0"

	tc := checker.NewToolChecker(cmd)
	result, err := tc.Check(context.Background(), "node", ">=18.0.0")

	require.NoError(t, err)
	assert.Equal(t, report.StatusOK, result.Status)
	assert.Equal(t, "20.11.0", result.FoundVersion)
	assert.Equal(t, ">=18.0.0", result.ExpectedRange)
}

func TestCheckTool_Installed_ConstraintNotSatisfied(t *testing.T) {
	cmd := NewMockCommander()
	cmd.Responses["node --version"] = "v16.0.0"

	tc := checker.NewToolChecker(cmd)
	result, err := tc.Check(context.Background(), "node", ">=18.0.0")

	require.NoError(t, err)
	assert.Equal(t, report.StatusOutdated, result.Status)
	assert.Equal(t, "16.0.0", result.FoundVersion)
	assert.Contains(t, result.Message, "does not satisfy")
}

func TestCheckTool_NotInstalled(t *testing.T) {
	cmd := NewMockCommander()
	cmd.Errors["docker version --format {{.Client.Version}}"] = errors.New("executable file not found in $PATH")

	tc := checker.NewToolChecker(cmd)
	result, err := tc.Check(context.Background(), "docker", "")

	require.NoError(t, err)
	assert.Equal(t, report.StatusMissing, result.Status)
	assert.Empty(t, result.FoundVersion)
	assert.Contains(t, result.Message, "not found")
}

func TestCheckTool_UnknownTool(t *testing.T) {
	cmd := NewMockCommander()
	tc := checker.NewToolChecker(cmd)
	_, err := tc.Check(context.Background(), "unknown-tool-xyz", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown tool")
}

func TestCheckTool_AllSupportedTools(t *testing.T) {
	// Verify that the tool registry includes all expected tools.
	supported := checker.SupportedTools()
	expected := []string{
		"go", "node", "npm", "docker", "git",
		"python3", "pip3", "make", "curl", "jq",
		"kubectl", "terraform", "helm", "rustc", "java",
	}
	for _, tool := range expected {
		assert.Contains(t, supported, tool, "tool %q should be in supported list", tool)
	}
}

func TestCheckTool_DockerVersion(t *testing.T) {
	cmd := NewMockCommander()
	cmd.Responses["docker version --format {{.Client.Version}}"] = "26.1.3"

	tc := checker.NewToolChecker(cmd)
	result, err := tc.Check(context.Background(), "docker", ">=24.0.0")

	require.NoError(t, err)
	assert.Equal(t, report.StatusOK, result.Status)
	assert.Equal(t, "26.1.3", result.FoundVersion)
}

func TestCheckTool_GitVersion(t *testing.T) {
	cmd := NewMockCommander()
	cmd.Responses["git --version"] = "git version 2.43.0"

	tc := checker.NewToolChecker(cmd)
	result, err := tc.Check(context.Background(), "git", ">=2.40.0")

	require.NoError(t, err)
	assert.Equal(t, report.StatusOK, result.Status)
	assert.Equal(t, "2.43.0", result.FoundVersion)
}

func TestCheckTool_PythonVersion(t *testing.T) {
	cmd := NewMockCommander()
	cmd.Responses["python3 --version"] = "Python 3.11.8"

	tc := checker.NewToolChecker(cmd)
	result, err := tc.Check(context.Background(), "python3", ">=3.10.0")

	require.NoError(t, err)
	assert.Equal(t, report.StatusOK, result.Status)
	assert.Equal(t, "3.11.8", result.FoundVersion)
}

func TestCheckTool_NpmVersion(t *testing.T) {
	cmd := NewMockCommander()
	cmd.Responses["npm --version"] = "10.2.4"

	tc := checker.NewToolChecker(cmd)
	result, err := tc.Check(context.Background(), "npm", ">=9.0.0")

	require.NoError(t, err)
	assert.Equal(t, report.StatusOK, result.Status)
	assert.Equal(t, "10.2.4", result.FoundVersion)
}

func TestCheckTool_KubectlVersion(t *testing.T) {
	cmd := NewMockCommander()
	cmd.Responses["kubectl version --client --short"] = "Client Version: v1.29.2"

	tc := checker.NewToolChecker(cmd)
	result, err := tc.Check(context.Background(), "kubectl", ">=1.28.0")

	require.NoError(t, err)
	assert.Equal(t, report.StatusOK, result.Status)
	assert.Equal(t, "1.29.2", result.FoundVersion)
}

func TestCheckTool_TerraformVersion(t *testing.T) {
	cmd := NewMockCommander()
	cmd.Responses["terraform version -json"] = `{"terraform_version":"1.7.4"}`

	tc := checker.NewToolChecker(cmd)
	result, err := tc.Check(context.Background(), "terraform", ">=1.5.0")

	require.NoError(t, err)
	assert.Equal(t, report.StatusOK, result.Status)
	assert.Equal(t, "1.7.4", result.FoundVersion)
}

func TestCheckTools_Multiple(t *testing.T) {
	cmd := NewMockCommander()
	cmd.Responses["go version"] = "go version go1.24.2 darwin/arm64"
	cmd.Responses["git --version"] = "git version 2.43.0"
	cmd.Errors["docker version --format {{.Client.Version}}"] = errors.New("not found")

	tc := checker.NewToolChecker(cmd)
	results, err := tc.CheckMultiple(context.Background(), []checker.ToolSpec{
		{Name: "go", Constraint: ">=1.21.0"},
		{Name: "git", Constraint: ""},
		{Name: "docker", Constraint: ">=24.0.0"},
	})

	require.NoError(t, err)
	require.Len(t, results, 3)
	assert.Equal(t, report.StatusOK, results[0].Status)
	assert.Equal(t, report.StatusOK, results[1].Status)
	assert.Equal(t, report.StatusMissing, results[2].Status)
}
