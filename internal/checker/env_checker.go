package checker

import (
	"fmt"
	"regexp"
	"sort"

	"github.com/pouyasadri/env-validator-mcp/internal/report"
)

// EnvSpec defines the validation criteria for a single environment variable.
type EnvSpec struct {
	// RequireNonEmpty means the variable must exist AND have a non-empty value.
	RequireNonEmpty bool
	// Pattern is an optional regular expression the value must match.
	// An empty string disables pattern checking.
	Pattern string
}

// EnvChecker validates environment variables using a pluggable EnvReader.
type EnvChecker struct {
	env EnvReader
}

// NewEnvChecker constructs an EnvChecker with the given EnvReader.
func NewEnvChecker(env EnvReader) *EnvChecker {
	return &EnvChecker{env: env}
}

// Check validates a single environment variable against an EnvSpec.
// Returns an error only for configuration problems (e.g. invalid regex), not for
// missing/misconfigured variables — those are reflected in the Result status.
func (ec *EnvChecker) Check(key string, spec EnvSpec) (report.Result, error) {
	// Validate the regex pattern up front so we can return a hard error.
	var rx *regexp.Regexp
	if spec.Pattern != "" {
		var err error
		rx, err = regexp.Compile(spec.Pattern)
		if err != nil {
			return report.Result{}, fmt.Errorf("invalid pattern for %q: %w", key, err)
		}
	}

	value, exists := ec.env.LookupEnv(key)
	if !exists {
		return report.Result{
			Name:    key,
			Status:  report.StatusMissing,
			Message: fmt.Sprintf("environment variable %s is not set", key),
		}, nil
	}

	if spec.RequireNonEmpty && value == "" {
		return report.Result{
			Name:    key,
			Status:  report.StatusMisconfigured,
			Message: fmt.Sprintf("environment variable %s is set but empty", key),
		}, nil
	}

	if rx != nil && !rx.MatchString(value) {
		return report.Result{
			Name:    key,
			Status:  report.StatusMisconfigured,
			Message: fmt.Sprintf("environment variable %s value does not match pattern %q", key, spec.Pattern),
		}, nil
	}

	return report.Result{
		Name:    key,
		Status:  report.StatusOK,
		Message: fmt.Sprintf("environment variable %s is set", key),
	}, nil
}

// CheckMultiple validates multiple environment variables concurrently-safe but
// deterministically ordered (sorted by key) to produce consistent output.
func (ec *EnvChecker) CheckMultiple(specs map[string]EnvSpec) []report.Result {
	// Sort keys for deterministic output.
	keys := make([]string, 0, len(specs))
	for k := range specs {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	results := make([]report.Result, 0, len(specs))
	for _, key := range keys {
		// Pattern errors are treated as misconfigured for batch runs — we don't
		// want one bad spec to abort the entire sweep.
		res, err := ec.Check(key, specs[key])
		if err != nil {
			res = report.Result{
				Name:    key,
				Status:  report.StatusError,
				Message: err.Error(),
			}
		}
		results = append(results, res)
	}
	return results
}
