package checker

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/pouyasadri/env-validator-mcp/internal/report"
)

// FileSpec defines the validation criteria for a single config file or directory.
type FileSpec struct {
	// ValidateJSON, if true, reads the file and checks that its content is valid JSON.
	ValidateJSON bool
	// RequiredKeys is a list of top-level JSON keys that must be present.
	// Only evaluated when ValidateJSON is true.
	RequiredKeys []string
}

// ConfigChecker validates the existence and optionally the content of config files.
type ConfigChecker struct {
	fs FileSystem
}

// NewConfigChecker constructs a ConfigChecker with the given FileSystem.
func NewConfigChecker(fs FileSystem) *ConfigChecker {
	return &ConfigChecker{fs: fs}
}

// Check validates a single file or directory path against a FileSpec.
func (cc *ConfigChecker) Check(path string, spec FileSpec) (report.Result, error) {
	_, err := cc.fs.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return report.Result{
				Name:    path,
				Status:  report.StatusMissing,
				Message: fmt.Sprintf("%s was not found", path),
			}, nil
		}
		return report.Result{
			Name:    path,
			Status:  report.StatusError,
			Message: fmt.Sprintf("could not stat %s: %v", path, err),
		}, nil
	}

	// File/dir exists — run optional content checks.
	if spec.ValidateJSON {
		return cc.validateJSON(path, spec)
	}

	return report.Result{
		Name:    path,
		Status:  report.StatusOK,
		Message: fmt.Sprintf("%s exists", path),
	}, nil
}

// validateJSON reads the file, checks JSON validity, and optionally verifies required top-level keys.
func (cc *ConfigChecker) validateJSON(path string, spec FileSpec) (report.Result, error) {
	content, err := cc.fs.ReadFile(path)
	if err != nil {
		return report.Result{
			Name:    path,
			Status:  report.StatusError,
			Message: fmt.Sprintf("could not read %s: %v", path, err),
		}, nil
	}

	var parsed map[string]any
	if err := json.Unmarshal(content, &parsed); err != nil {
		return report.Result{
			Name:    path,
			Status:  report.StatusMisconfigured,
			Message: fmt.Sprintf("%s contains invalid JSON: %v", path, err),
		}, nil
	}

	if len(spec.RequiredKeys) > 0 {
		missing := make([]string, 0)
		for _, key := range spec.RequiredKeys {
			if _, ok := parsed[key]; !ok {
				missing = append(missing, key)
			}
		}
		if len(missing) > 0 {
			return report.Result{
				Name:    path,
				Status:  report.StatusMisconfigured,
				Message: fmt.Sprintf("%s is missing required keys: %s", path, strings.Join(missing, ", ")),
			}, nil
		}
	}

	return report.Result{
		Name:    path,
		Status:  report.StatusOK,
		Message: fmt.Sprintf("%s exists and is valid JSON", path),
	}, nil
}

// CheckMultiple validates multiple file paths and returns results sorted by path.
func (cc *ConfigChecker) CheckMultiple(specs map[string]FileSpec) []report.Result {
	// Sort paths for deterministic output.
	paths := make([]string, 0, len(specs))
	for p := range specs {
		paths = append(paths, p)
	}
	sort.Strings(paths)

	results := make([]report.Result, 0, len(specs))
	for _, path := range paths {
		res, err := cc.Check(path, specs[path])
		if err != nil {
			res = report.Result{
				Name:    path,
				Status:  report.StatusError,
				Message: err.Error(),
			}
		}
		results = append(results, res)
	}
	return results
}
