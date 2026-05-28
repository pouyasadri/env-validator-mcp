package checker

import (
	"bytes"
	"context"
	"os"
	"os/exec"
)

// RealCommander implements Commander using the OS process model.
// It combines stdout and stderr since many tools (e.g. java -version) write to stderr.
type RealCommander struct{}

// NewRealCommander returns the production Commander implementation.
func NewRealCommander() *RealCommander {
	return &RealCommander{}
}

// Run executes the named program with the given arguments and returns combined output.
func (r *RealCommander) Run(ctx context.Context, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf // capture stderr too (java -version, etc.)
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// RealEnvReader implements EnvReader using the real process environment.
type RealEnvReader struct{}

// NewRealEnvReader returns the production EnvReader.
func NewRealEnvReader() *RealEnvReader { return &RealEnvReader{} }

// LookupEnv wraps os.LookupEnv.
func (r *RealEnvReader) LookupEnv(key string) (string, bool) {
	return os.LookupEnv(key)
}

// RealFileSystem implements FileSystem using the real OS file system.
type RealFileSystem struct{}

// NewRealFileSystem returns the production FileSystem.
func NewRealFileSystem() *RealFileSystem { return &RealFileSystem{} }

// Stat wraps os.Stat.
func (r *RealFileSystem) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

// ReadFile wraps os.ReadFile.
func (r *RealFileSystem) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(name)
}
