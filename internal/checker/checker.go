// Package checker defines interfaces and implementations for inspecting the
// local development environment. All external effects (command execution,
// environment variable reads, file system access) are abstracted behind
// interfaces to enable dependency injection and thorough TDD.
package checker

import (
	"context"
	"os"
)

// Commander executes external commands and returns their combined stdout output.
// In production, RealCommander is used. In tests, MockCommander is injected.
type Commander interface {
	Run(ctx context.Context, name string, args ...string) (stdout string, err error)
}

// EnvReader abstracts os.LookupEnv so environment variable reads can be
// controlled in tests without mutating the real process environment.
type EnvReader interface {
	LookupEnv(key string) (string, bool)
}

// FileSystem abstracts file system operations needed by the config checker.
type FileSystem interface {
	Stat(name string) (os.FileInfo, error)
	ReadFile(name string) ([]byte, error)
}
