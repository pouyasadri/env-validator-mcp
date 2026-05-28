// Package semver provides utilities for parsing, normalizing, and comparing
// semantic version strings encountered in real-world CLI tool output.
package semver

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	gover "golang.org/x/mod/semver"
)

// versionPattern matches version strings in the form X.Y.Z or X.Y embedded in
// arbitrary text produced by CLI tools. It handles forms like:
//   - "go1.24.2"        → 1.24.2
//   - "v20.11.0"        → 20.11.0
//   - "version 2.43.0" → 2.43.0
//   - "26.1.3,"         → 26.1.3
//
// The key insight: we require that the matched number is NOT preceded by another
// digit (which would mean we are in the middle of a longer number). Using a
// negative lookbehind equivalent via anchoring:
// We match an optional non-digit boundary before the first digit group.
var versionPattern = regexp.MustCompile(`(?:^|[^\d])(\d+)\.(\d+)(?:\.(\d+))?(?:[^\d]|$)`)

// ExtractVersion extracts the first semantic version number (X.Y.Z or X.Y)
// from a raw string such as tool --version output.
func ExtractVersion(raw string) (string, error) {
	if raw == "" {
		return "", fmt.Errorf("empty version string")
	}

	m := versionPattern.FindStringSubmatch(raw)
	if m == nil {
		return "", fmt.Errorf("no version number found in %q", raw)
	}

	// m[0]=full match, m[1]=major, m[2]=minor, m[3]=patch (may be empty)
	major := m[1]
	minor := m[2]
	patch := "0"
	if m[3] != "" {
		patch = m[3]
	}

	return fmt.Sprintf("%s.%s.%s", major, minor, patch), nil
}

// Normalize converts a raw version string (with or without "v" prefix, with optional
// build metadata) into a canonical vX.Y.Z string compatible with golang.org/x/mod/semver.
func Normalize(raw string) (string, error) {
	if raw == "" {
		return "", fmt.Errorf("empty version string")
	}

	// Strip leading "v"
	cleaned := strings.TrimPrefix(raw, "v")

	// Strip build metadata (everything after "+")
	if idx := strings.Index(cleaned, "+"); idx != -1 {
		cleaned = cleaned[:idx]
	}

	// Split into semver components, keep pre-release suffix
	preRelease := ""
	if idx := strings.Index(cleaned, "-"); idx != -1 {
		preRelease = cleaned[idx:] // e.g. "-rc1"
		cleaned = cleaned[:idx]
	}

	parts := strings.Split(cleaned, ".")
	switch len(parts) {
	case 2:
		parts = append(parts, "0")
	case 3:
		// already fine
	default:
		return "", fmt.Errorf("cannot parse version %q: expected X.Y or X.Y.Z format", raw)
	}

	// Validate that each part is a non-negative integer
	for _, p := range parts {
		if _, err := strconv.Atoi(p); err != nil {
			return "", fmt.Errorf("cannot parse version %q: part %q is not an integer", raw, p)
		}
	}

	canonical := "v" + strings.Join(parts, ".") + preRelease

	if !gover.IsValid(canonical) {
		return "", fmt.Errorf("cannot parse version %q: not valid semver", raw)
	}

	return canonical, nil
}

// Satisfies reports whether version satisfies the given constraint string.
// Supported operators: >=, >, <=, <, =
// If constraint is empty, any version satisfies it.
func Satisfies(version, constraint string) (bool, error) {
	if constraint == "" {
		return true, nil
	}

	// Parse operator and target version
	op, target, err := parseConstraint(constraint)
	if err != nil {
		return false, err
	}

	v, err := Normalize(version)
	if err != nil {
		return false, fmt.Errorf("version: %w", err)
	}

	t, err := Normalize(target)
	if err != nil {
		return false, fmt.Errorf("constraint version: %w", err)
	}

	cmp := gover.Compare(v, t)

	switch op {
	case ">=":
		return cmp >= 0, nil
	case ">":
		return cmp > 0, nil
	case "<=":
		return cmp <= 0, nil
	case "<":
		return cmp < 0, nil
	case "=", "==":
		return cmp == 0, nil
	default:
		return false, fmt.Errorf("unsupported constraint operator %q", op)
	}
}

// parseConstraint splits a constraint like ">=1.21.0" into operator ">=", version "1.21.0".
func parseConstraint(c string) (op, version string, err error) {
	for _, prefix := range []string{">=", "<=", ">", "<", "==", "="} {
		if strings.HasPrefix(c, prefix) {
			return prefix, strings.TrimSpace(c[len(prefix):]), nil
		}
	}
	return "", "", fmt.Errorf("constraint %q has no recognized operator (use >=, >, <=, <, =)", c)
}
