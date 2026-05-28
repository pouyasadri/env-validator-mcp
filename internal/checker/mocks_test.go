// Package checker_test provides shared test helpers and mocks for the checker package.
package checker_test

import (
	"context"
	"os"
	"time"
)

// --- MockCommander ---

// MockCommander is a test double for checker.Commander.
// Register expected calls in Responses before running code under test.
type MockCommander struct {
	// Responses maps "name args[0] args[1]..." → stdout string.
	Responses map[string]string
	// Errors maps the same key → error to return.
	Errors map[string]error
	// Calls records every invocation for assertion.
	Calls []MockCall
}

// MockCall records a single invocation of MockCommander.Run.
type MockCall struct {
	Name string
	Args []string
}

func NewMockCommander() *MockCommander {
	return &MockCommander{
		Responses: make(map[string]string),
		Errors:    make(map[string]error),
	}
}

func (m *MockCommander) Run(_ context.Context, name string, args ...string) (string, error) {
	m.Calls = append(m.Calls, MockCall{Name: name, Args: args})
	key := cmdKey(name, args)
	if err, ok := m.Errors[key]; ok {
		return "", err
	}
	return m.Responses[key], nil
}

func cmdKey(name string, args []string) string {
	key := name
	for _, a := range args {
		key += " " + a
	}
	return key
}

// --- MockEnvReader ---

// MockEnvReader is a test double for checker.EnvReader.
type MockEnvReader struct {
	Vars map[string]string
}

func NewMockEnvReader(vars map[string]string) *MockEnvReader {
	return &MockEnvReader{Vars: vars}
}

func (m *MockEnvReader) LookupEnv(key string) (string, bool) {
	v, ok := m.Vars[key]
	return v, ok
}

// --- MockFileSystem ---

// MockFileInfo is a minimal os.FileInfo implementation for tests.
type MockFileInfo struct {
	name    string
	size    int64
	isDir   bool
	modTime time.Time
}

func NewMockFileInfo(name string, isDir bool) *MockFileInfo {
	return &MockFileInfo{name: name, isDir: isDir}
}

func (m *MockFileInfo) Name() string       { return m.name }
func (m *MockFileInfo) Size() int64        { return m.size }
func (m *MockFileInfo) IsDir() bool        { return m.isDir }
func (m *MockFileInfo) ModTime() time.Time { return m.modTime }
func (m *MockFileInfo) Mode() os.FileMode  { return 0644 }
func (m *MockFileInfo) Sys() any           { return nil }

// MockFileSystem is a test double for checker.FileSystem.
type MockFileSystem struct {
	Files  map[string][]byte // path → content
	Dirs   map[string]bool   // path → isDir
	Errors map[string]error  // path → stat/read error
}

func NewMockFileSystem() *MockFileSystem {
	return &MockFileSystem{
		Files:  make(map[string][]byte),
		Dirs:   make(map[string]bool),
		Errors: make(map[string]error),
	}
}

func (m *MockFileSystem) AddFile(path, content string) {
	m.Files[path] = []byte(content)
}

func (m *MockFileSystem) AddDir(path string) {
	m.Dirs[path] = true
}

func (m *MockFileSystem) Stat(name string) (os.FileInfo, error) {
	if err, ok := m.Errors[name]; ok {
		return nil, err
	}
	if _, ok := m.Files[name]; ok {
		return NewMockFileInfo(name, false), nil
	}
	if m.Dirs[name] {
		return NewMockFileInfo(name, true), nil
	}
	return nil, os.ErrNotExist
}

func (m *MockFileSystem) ReadFile(name string) ([]byte, error) {
	if err, ok := m.Errors[name]; ok {
		return nil, err
	}
	if content, ok := m.Files[name]; ok {
		return content, nil
	}
	return nil, os.ErrNotExist
}
