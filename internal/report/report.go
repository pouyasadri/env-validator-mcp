// Package report defines the core data types shared across all environment checkers.
package report

import "fmt"

// Status represents the outcome of a single environment check.
type Status string

const (
	// StatusOK means the check passed — the tool/env/file exists and meets requirements.
	StatusOK Status = "ok"
	// StatusMissing means the tool, environment variable, or file was not found.
	StatusMissing Status = "missing"
	// StatusOutdated means the tool is installed but its version does not satisfy the constraint.
	StatusOutdated Status = "outdated"
	// StatusMisconfigured means the item exists but its value/content is not valid.
	StatusMisconfigured Status = "misconfigured"
	// StatusError means an unexpected error occurred during the check itself.
	StatusError Status = "error"
)

// Result is the outcome of a single environment check.
type Result struct {
	// Name is the identifier of the checked item (e.g. "go", "DATABASE_URL", ".env").
	Name string `json:"name"`
	// Status is the outcome of the check.
	Status Status `json:"status"`
	// FoundVersion is the version string detected for tool checks. Empty for non-version checks.
	FoundVersion string `json:"found_version,omitempty"`
	// ExpectedRange is the semver constraint provided by the caller. Empty when no constraint was given.
	ExpectedRange string `json:"expected_range,omitempty"`
	// Message is a human-readable explanation of the check outcome.
	Message string `json:"message"`
}

// Report aggregates multiple Results from a full environment validation sweep.
type Report struct {
	Results    []Result `json:"results"`
	Summary    string   `json:"summary"`
	OKCount    int      `json:"ok_count"`
	IssueCount int      `json:"issue_count"`
}

// NewReport constructs a Report from a slice of Results, computing counts and a summary.
func NewReport(results []Result) Report {
	if results == nil {
		results = []Result{}
	}

	ok, issues := 0, 0
	for _, r := range results {
		if r.Status == StatusOK {
			ok++
		} else {
			issues++
		}
	}

	var summary string
	switch {
	case issues == 0 && ok > 0:
		summary = fmt.Sprintf("All checks passed (%d OK)", ok)
	case issues == 0 && ok == 0:
		summary = "No checks were run"
	default:
		summary = fmt.Sprintf("%d OK, %d issue(s) found", ok, issues)
	}

	return Report{
		Results:    results,
		Summary:    summary,
		OKCount:    ok,
		IssueCount: issues,
	}
}
