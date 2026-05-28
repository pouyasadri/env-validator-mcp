package semver_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pouyasadri/env-validator-mcp/internal/semver"
)

func TestNormalize(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "already canonical", input: "v1.24.2", want: "v1.24.2"},
		{name: "without v prefix", input: "1.24.2", want: "v1.24.2"},
		{name: "two parts", input: "1.21", want: "v1.21.0"},
		{name: "with build metadata stripped", input: "1.21.0+build123", want: "v1.21.0"},
		{name: "with pre-release kept", input: "1.21.0-rc1", want: "v1.21.0-rc1"},
		{name: "empty string", input: "", wantErr: true},
		{name: "non-numeric", input: "abc", wantErr: true},
		{name: "go style major only", input: "1", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := semver.Normalize(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSatisfies(t *testing.T) {
	tests := []struct {
		name       string
		version    string
		constraint string
		want       bool
		wantErr    bool
	}{
		// Greater-or-equal
		{name: "gte satisfied", version: "1.24.0", constraint: ">=1.21.0", want: true},
		{name: "gte equal", version: "1.21.0", constraint: ">=1.21.0", want: true},
		{name: "gte not satisfied", version: "1.20.0", constraint: ">=1.21.0", want: false},

		// Greater-than
		{name: "gt satisfied", version: "1.22.0", constraint: ">1.21.0", want: true},
		{name: "gt equal fails", version: "1.21.0", constraint: ">1.21.0", want: false},

		// Less-or-equal
		{name: "lte satisfied", version: "1.20.0", constraint: "<=1.21.0", want: true},
		{name: "lte equal", version: "1.21.0", constraint: "<=1.21.0", want: true},
		{name: "lte not satisfied", version: "1.22.0", constraint: "<=1.21.0", want: false},

		// Less-than
		{name: "lt satisfied", version: "1.20.0", constraint: "<1.21.0", want: true},
		{name: "lt equal fails", version: "1.21.0", constraint: "<1.21.0", want: false},

		// Exact
		{name: "exact match", version: "1.21.0", constraint: "=1.21.0", want: true},
		{name: "exact no match", version: "1.22.0", constraint: "=1.21.0", want: false},

		// No constraint (any version)
		{name: "empty constraint always ok", version: "1.0.0", constraint: "", want: true},

		// Invalid inputs
		{name: "invalid version", version: "not-a-version", constraint: ">=1.0.0", wantErr: true},
		{name: "invalid constraint", version: "1.0.0", constraint: ">>1.0.0", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := semver.Satisfies(tt.version, tt.constraint)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestExtractVersion(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "plain version", input: "1.24.2", want: "1.24.2"},
		{name: "go style", input: "go version go1.24.2 darwin/arm64", want: "1.24.2"},
		{name: "node style", input: "v20.11.0", want: "20.11.0"},
		{name: "docker style", input: "Docker version 26.1.3, build b72abbb", want: "26.1.3"},
		{name: "git style", input: "git version 2.43.0", want: "2.43.0"},
		{name: "python style", input: "Python 3.11.8", want: "3.11.8"},
		{name: "npm style", input: "10.2.4", want: "10.2.4"},
		{name: "kubectl style", input: "Client Version: v1.29.2", want: "1.29.2"},
		{name: "terraform style", input: "Terraform v1.7.4\non linux_arm64", want: "1.7.4"},
		{name: "empty string", input: "", wantErr: true},
		{name: "no version found", input: "some text with no version", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := semver.ExtractVersion(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
