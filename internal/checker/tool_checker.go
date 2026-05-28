package checker

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pouyasadri/env-validator-mcp/internal/report"
	"github.com/pouyasadri/env-validator-mcp/internal/semver"
)

// toolConfig defines how to invoke a specific CLI tool and extract its version.
type toolConfig struct {
	// args are the command-line arguments to pass to the tool binary.
	args []string
	// extractFn extracts a clean X.Y.Z version string from the raw command output.
	extractFn func(output string) (string, error)
}

// toolRegistry maps tool names to their invocation configuration.
var toolRegistry = map[string]toolConfig{
	"go": {
		args:      []string{"version"},
		extractFn: extractViaRegex,
	},
	"node": {
		args:      []string{"--version"},
		extractFn: extractViaRegex,
	},
	"npm": {
		args:      []string{"--version"},
		extractFn: extractViaRegex,
	},
	"docker": {
		args:      []string{"version", "--format", "{{.Client.Version}}"},
		extractFn: extractViaRegex,
	},
	"git": {
		args:      []string{"--version"},
		extractFn: extractViaRegex,
	},
	"python3": {
		args:      []string{"--version"},
		extractFn: extractViaRegex,
	},
	"pip3": {
		args:      []string{"--version"},
		extractFn: extractViaRegex,
	},
	"make": {
		args:      []string{"--version"},
		extractFn: extractViaRegex,
	},
	"curl": {
		args:      []string{"--version"},
		extractFn: extractViaRegex,
	},
	"jq": {
		args:      []string{"--version"},
		extractFn: extractViaRegex,
	},
	"kubectl": {
		args:      []string{"version", "--client", "--short"},
		extractFn: extractViaRegex,
	},
	"terraform": {
		args:      []string{"version", "-json"},
		extractFn: extractTerraformVersion,
	},
	"helm": {
		args:      []string{"version", "--short"},
		extractFn: extractViaRegex,
	},
	"rustc": {
		args:      []string{"--version"},
		extractFn: extractViaRegex,
	},
	"java": {
		args:      []string{"-version"},
		extractFn: extractViaRegex,
	},
}

// extractViaRegex delegates to the semver package's general-purpose extractor.
func extractViaRegex(output string) (string, error) {
	return semver.ExtractVersion(output)
}

// extractTerraformVersion parses the JSON output of `terraform version -json`.
func extractTerraformVersion(output string) (string, error) {
	// terraform version -json outputs: {"terraform_version":"1.7.4", ...}
	var data struct {
		TerraformVersion string `json:"terraform_version"`
	}
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &data); err == nil && data.TerraformVersion != "" {
		return data.TerraformVersion, nil
	}
	// Fall back to regex if JSON parse fails (older terraform versions)
	return semver.ExtractVersion(output)
}

// SupportedTools returns the sorted list of tool names this checker knows about.
func SupportedTools() []string {
	tools := make([]string, 0, len(toolRegistry))
	for name := range toolRegistry {
		tools = append(tools, name)
	}
	return tools
}

// ToolSpec bundles a tool name with an optional semver constraint.
type ToolSpec struct {
	Name       string
	Constraint string
}

// ToolChecker checks whether CLI tools are installed and meet version constraints.
type ToolChecker struct {
	cmd Commander
}

// NewToolChecker constructs a ToolChecker with the given Commander.
func NewToolChecker(cmd Commander) *ToolChecker {
	return &ToolChecker{cmd: cmd}
}

// Check verifies a single tool.
// constraint may be empty (any version passes), or a semver constraint like ">=1.21.0".
func (tc *ToolChecker) Check(ctx context.Context, toolName, constraint string) (report.Result, error) {
	cfg, ok := toolRegistry[toolName]
	if !ok {
		return report.Result{}, fmt.Errorf("unknown tool %q: use SupportedTools() to list valid names", toolName)
	}

	stdout, err := tc.cmd.Run(ctx, toolName, cfg.args...)
	if err != nil {
		return report.Result{
			Name:    toolName,
			Status:  report.StatusMissing,
			Message: fmt.Sprintf("%s is not found or not executable: %v", toolName, err),
		}, nil
	}

	rawVersion, err := cfg.extractFn(stdout)
	if err != nil {
		return report.Result{
			Name:    toolName,
			Status:  report.StatusError,
			Message: fmt.Sprintf("could not extract version from %s output: %v", toolName, err),
		}, nil
	}

	if constraint == "" {
		return report.Result{
			Name:         toolName,
			Status:       report.StatusOK,
			FoundVersion: rawVersion,
			Message:      fmt.Sprintf("%s is installed (version %s)", toolName, rawVersion),
		}, nil
	}

	ok, err = semver.Satisfies(rawVersion, constraint)
	if err != nil {
		return report.Result{
			Name:          toolName,
			Status:        report.StatusError,
			FoundVersion:  rawVersion,
			ExpectedRange: constraint,
			Message:       fmt.Sprintf("version comparison failed: %v", err),
		}, nil
	}

	if ok {
		return report.Result{
			Name:          toolName,
			Status:        report.StatusOK,
			FoundVersion:  rawVersion,
			ExpectedRange: constraint,
			Message:       fmt.Sprintf("%s %s is installed and satisfies %s", toolName, rawVersion, constraint),
		}, nil
	}

	return report.Result{
		Name:          toolName,
		Status:        report.StatusOutdated,
		FoundVersion:  rawVersion,
		ExpectedRange: constraint,
		Message:       fmt.Sprintf("%s %s does not satisfy constraint %s", toolName, rawVersion, constraint),
	}, nil
}

// CheckMultiple runs Check for each ToolSpec and collects results.
// It never short-circuits on individual failures — all specs are checked.
func (tc *ToolChecker) CheckMultiple(ctx context.Context, specs []ToolSpec) ([]report.Result, error) {
	results := make([]report.Result, 0, len(specs))
	for _, spec := range specs {
		res, err := tc.Check(ctx, spec.Name, spec.Constraint)
		if err != nil {
			// Unknown tool — return immediately so callers know about config errors.
			return nil, err
		}
		results = append(results, res)
	}
	return results, nil
}
