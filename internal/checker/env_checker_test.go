package checker_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pouyasadri/env-validator-mcp/internal/checker"
	"github.com/pouyasadri/env-validator-mcp/internal/report"
)

func TestCheckEnvVar_Exists_NoPattern(t *testing.T) {
	env := NewMockEnvReader(map[string]string{"HOME": "/Users/test"})
	ec := checker.NewEnvChecker(env)

	result, err := ec.Check("HOME", checker.EnvSpec{})
	require.NoError(t, err)
	assert.Equal(t, report.StatusOK, result.Status)
	assert.Equal(t, "HOME", result.Name)
	assert.Contains(t, result.Message, "is set")
}

func TestCheckEnvVar_Missing(t *testing.T) {
	env := NewMockEnvReader(map[string]string{})
	ec := checker.NewEnvChecker(env)

	result, err := ec.Check("DATABASE_URL", checker.EnvSpec{})
	require.NoError(t, err)
	assert.Equal(t, report.StatusMissing, result.Status)
	assert.Contains(t, result.Message, "is not set")
}

func TestCheckEnvVar_ExistsButEmpty_RequiredNonEmpty(t *testing.T) {
	env := NewMockEnvReader(map[string]string{"API_KEY": ""})
	ec := checker.NewEnvChecker(env)

	result, err := ec.Check("API_KEY", checker.EnvSpec{RequireNonEmpty: true})
	require.NoError(t, err)
	assert.Equal(t, report.StatusMisconfigured, result.Status)
	assert.Contains(t, result.Message, "empty")
}

func TestCheckEnvVar_ExistsEmpty_NotRequired(t *testing.T) {
	env := NewMockEnvReader(map[string]string{"OPTIONAL_VAR": ""})
	ec := checker.NewEnvChecker(env)

	result, err := ec.Check("OPTIONAL_VAR", checker.EnvSpec{})
	require.NoError(t, err)
	assert.Equal(t, report.StatusOK, result.Status)
}

func TestCheckEnvVar_PatternMatch_Passes(t *testing.T) {
	env := NewMockEnvReader(map[string]string{"PORT": "8080"})
	ec := checker.NewEnvChecker(env)

	result, err := ec.Check("PORT", checker.EnvSpec{Pattern: `^\d+$`})
	require.NoError(t, err)
	assert.Equal(t, report.StatusOK, result.Status)
}

func TestCheckEnvVar_PatternMatch_Fails(t *testing.T) {
	env := NewMockEnvReader(map[string]string{"PORT": "not-a-port"})
	ec := checker.NewEnvChecker(env)

	result, err := ec.Check("PORT", checker.EnvSpec{Pattern: `^\d+$`})
	require.NoError(t, err)
	assert.Equal(t, report.StatusMisconfigured, result.Status)
	assert.Contains(t, result.Message, "does not match")
}

func TestCheckEnvVar_InvalidPattern(t *testing.T) {
	env := NewMockEnvReader(map[string]string{"X": "val"})
	ec := checker.NewEnvChecker(env)

	_, err := ec.Check("X", checker.EnvSpec{Pattern: `[invalid(`})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid pattern")
}

func TestCheckEnvVars_Multiple(t *testing.T) {
	env := NewMockEnvReader(map[string]string{
		"HOME":   "/Users/test",
		"GOPATH": "/Users/test/go",
	})
	ec := checker.NewEnvChecker(env)

	specs := map[string]checker.EnvSpec{
		"HOME":        {},
		"GOPATH":      {},
		"MISSING_VAR": {},
	}

	results := ec.CheckMultiple(specs)
	require.Len(t, results, 3)

	statusByName := make(map[string]report.Status)
	for _, r := range results {
		statusByName[r.Name] = r.Status
	}
	assert.Equal(t, report.StatusOK, statusByName["HOME"])
	assert.Equal(t, report.StatusOK, statusByName["GOPATH"])
	assert.Equal(t, report.StatusMissing, statusByName["MISSING_VAR"])
}
