package report_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pouyasadri/env-validator-mcp/internal/report"
)

func TestStatus_Constants(t *testing.T) {
	assert.Equal(t, report.Status("ok"), report.StatusOK)
	assert.Equal(t, report.Status("missing"), report.StatusMissing)
	assert.Equal(t, report.Status("outdated"), report.StatusOutdated)
	assert.Equal(t, report.Status("misconfigured"), report.StatusMisconfigured)
	assert.Equal(t, report.Status("error"), report.StatusError)
}

func TestResult_JSONRoundTrip(t *testing.T) {
	r := report.Result{
		Name:          "go",
		Status:        report.StatusOK,
		FoundVersion:  "1.24.2",
		ExpectedRange: ">=1.21.0",
		Message:       "go is installed and meets the version requirement",
	}

	data, err := json.Marshal(r)
	require.NoError(t, err)

	var got report.Result
	require.NoError(t, json.Unmarshal(data, &got))
	assert.Equal(t, r, got)
}

func TestResult_OmitEmptyFields(t *testing.T) {
	r := report.Result{
		Name:    "git",
		Status:  report.StatusMissing,
		Message: "git is not installed",
	}

	data, err := json.Marshal(r)
	require.NoError(t, err)

	var m map[string]any
	require.NoError(t, json.Unmarshal(data, &m))

	assert.NotContains(t, m, "found_version", "found_version should be omitted when empty")
	assert.NotContains(t, m, "expected_range", "expected_range should be omitted when empty")
}

func TestNewReport_Counts(t *testing.T) {
	results := []report.Result{
		{Name: "go", Status: report.StatusOK, Message: "ok"},
		{Name: "node", Status: report.StatusMissing, Message: "missing"},
		{Name: "docker", Status: report.StatusOutdated, Message: "outdated"},
		{Name: "HOME", Status: report.StatusOK, Message: "ok"},
		{Name: "DATABASE_URL", Status: report.StatusMisconfigured, Message: "bad"},
	}

	r := report.NewReport(results)

	assert.Equal(t, 2, r.OKCount)
	assert.Equal(t, 3, r.IssueCount)
	assert.Len(t, r.Results, 5)
	assert.Contains(t, r.Summary, "2 OK")
	assert.Contains(t, r.Summary, "3 issue")
}

func TestNewReport_AllOK(t *testing.T) {
	results := []report.Result{
		{Name: "go", Status: report.StatusOK, Message: "ok"},
		{Name: "git", Status: report.StatusOK, Message: "ok"},
	}

	r := report.NewReport(results)

	assert.Equal(t, 2, r.OKCount)
	assert.Equal(t, 0, r.IssueCount)
	assert.Contains(t, r.Summary, "All checks passed")
}

func TestNewReport_Empty(t *testing.T) {
	r := report.NewReport(nil)
	assert.Equal(t, 0, r.OKCount)
	assert.Equal(t, 0, r.IssueCount)
	assert.Empty(t, r.Results)
}

func TestReport_JSONRoundTrip(t *testing.T) {
	results := []report.Result{
		{Name: "go", Status: report.StatusOK, Message: "all good"},
	}
	r := report.NewReport(results)

	data, err := json.Marshal(r)
	require.NoError(t, err)

	var got report.Report
	require.NoError(t, json.Unmarshal(data, &got))
	assert.Equal(t, r, got)
}
