package mcpserver

import (
	"context"
	"errors"
	"os"
	"time"
)

// TestDependencies returns a Dependencies struct with predictable mock implementations
// suitable for integration tests. The mocks simulate a real (but controlled) environment:
//
//   - "go" is installed at version 1.24.2
//   - "node" is installed at version 20.11.0
//   - "docker" is NOT installed (simulates missing tool)
//   - HOME and PATH are set; DEFINITELY_NOT_SET_XYZ_123 is absent
//   - /tmp directory exists; all other paths are absent
func TestDependencies() Dependencies {
	return Dependencies{
		Commander:  &testCommander{},
		EnvReader:  &testEnvReader{},
		FileSystem: &testFileSystem{},
	}
}

// testCommander simulates tool invocations for testing.
type testCommander struct{}

func (c *testCommander) Run(_ context.Context, name string, args ...string) (string, error) {
	key := name
	for _, a := range args {
		key += " " + a
	}
	responses := map[string]string{
		"go version":                       "go version go1.24.2 darwin/arm64",
		"node --version":                   "v20.11.0",
		"npm --version":                    "10.2.4",
		"git --version":                    "git version 2.43.0",
		"python3 --version":                "Python 3.11.8",
		"pip3 --version":                   "pip 23.3 from /usr/lib/python3/dist-packages/pip (python 3.11)",
		"make --version":                   "GNU Make 4.3",
		"curl --version":                   "curl 7.88.1 (x86_64-pc-linux-gnu)",
		"jq --version":                     "jq-1.6",
		"kubectl version --client --short": "Client Version: v1.29.2",
		"terraform version -json":          `{"terraform_version":"1.7.4"}`,
		"helm version --short":             "v3.14.0+g3fc9f4b",
		"rustc --version":                  "rustc 1.76.0 (07dca489a 2024-02-04)",
		"java -version":                    "openjdk version \"21.0.2\" 2024-01-16",
	}
	if resp, ok := responses[key]; ok {
		return resp, nil
	}
	// Simulate "not found" for everything else (e.g. docker)
	return "", errors.New("executable file not found in $PATH")
}

// testEnvReader provides a small controlled environment.
type testEnvReader struct{}

func (r *testEnvReader) LookupEnv(key string) (string, bool) {
	env := map[string]string{
		"HOME": "/Users/testuser",
		"PATH": "/usr/local/bin:/usr/bin:/bin",
	}
	v, ok := env[key]
	return v, ok
}

// testFileSystem simulates a minimal file system.
type testFileSystem struct{}

func (f *testFileSystem) Stat(name string) (os.FileInfo, error) {
	// Only /tmp exists in our test world.
	if name == "/tmp" {
		return testFileInfo{name: "/tmp", isDir: true}, nil
	}
	return nil, os.ErrNotExist
}

func (f *testFileSystem) ReadFile(name string) ([]byte, error) {
	return nil, os.ErrNotExist
}

type testFileInfo struct {
	name  string
	isDir bool
}

func (fi testFileInfo) Name() string       { return fi.name }
func (fi testFileInfo) Size() int64        { return 0 }
func (fi testFileInfo) IsDir() bool        { return fi.isDir }
func (fi testFileInfo) Mode() os.FileMode  { return 0755 }
func (fi testFileInfo) ModTime() time.Time { return time.Time{} }
func (fi testFileInfo) Sys() any           { return nil }
