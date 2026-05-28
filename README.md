# env-validator-mcp

[![Go](https://img.shields.io/badge/Go-1.24-00ADD8?logo=go)](https://go.dev/)
[![MCP SDK](https://img.shields.io/badge/MCP%20SDK-v1.6.1-blueviolet)](https://github.com/modelcontextprotocol/go-sdk)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

An **MCP (Model Context Protocol) server** that inspects your local development environment and reports missing dependencies, version mismatches, and misconfigured environment variables to any MCP-compatible AI assistant.

Works with **Claude Desktop**, **Cursor**, **Gemini CLI**, and any other client that supports stdio-transport MCP servers.

---

## Features

| MCP Tool | What it does |
|---|---|
| `check_tool_version` | Checks if a CLI tool is installed; optionally validates a semver constraint (e.g. `>=1.21.0`) |
| `check_env_vars` | Verifies environment variables are set, non-empty, and/or match a regex pattern |
| `check_config_files` | Confirms files/directories exist; optionally validates JSON files and checks for required keys |
| `validate_environment` | Full sweep: tools + env vars + config files, returns a structured JSON report |
| `list_checks` | Lists everything this server can inspect |

**Supported CLI tools:** `go`, `node`, `npm`, `docker`, `git`, `python3`, `pip3`, `make`, `curl`, `jq`, `kubectl`, `terraform`, `helm`, `rustc`, `java`

---

## Quick Start

### 1. Build

```bash
git clone https://github.com/pouyasadri/env-validator-mcp.git
cd env-validator-mcp
make build
# → bin/env-validator-mcp
```

### 2. Configure Claude Desktop

Add to `~/Library/Application Support/Claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "env-validator": {
      "command": "/Users/pouyasadri/Desktop/Projects/env-validator-mcp/bin/env-validator-mcp"
    }
  }
}
```

Restart Claude Desktop. You can now ask:

> *"Check if Go, Docker and Node are installed and meeting my minimum version requirements"*

> *"Is my DATABASE_URL environment variable set?"*

> *"Run a full environment validation and tell me what's missing"*

### 3. Configure Cursor / other MCP clients

Most clients support the same JSON config format. Point `command` at the binary path.

---

## Example Responses

### `validate_environment`

```json
{
  "results": [
    {
      "name": "go",
      "status": "ok",
      "found_version": "1.24.2",
      "expected_range": ">=1.21.0",
      "message": "go 1.24.2 is installed and satisfies >=1.21.0"
    },
    {
      "name": "docker",
      "status": "missing",
      "message": "docker is not found or not executable: executable file not found in $PATH"
    },
    {
      "name": "HOME",
      "status": "ok",
      "message": "environment variable HOME is set"
    }
  ],
  "summary": "6 OK, 2 issue(s) found",
  "ok_count": 6,
  "issue_count": 2
}
```

### `check_tool_version`

```json
// Arguments: { "tool": "node", "constraint": ">=18.0.0" }
{
  "name": "node",
  "status": "outdated",
  "found_version": "16.0.0",
  "expected_range": ">=18.0.0",
  "message": "node 16.0.0 does not satisfy constraint >=18.0.0"
}
```

---

## Semver Constraint Syntax

| Operator | Example | Meaning |
|---|---|---|
| `>=` | `>=1.21.0` | Version must be 1.21.0 or newer |
| `>` | `>1.20.0` | Version must be strictly newer than 1.20.0 |
| `<=` | `<=3.0.0` | Version must be 3.0.0 or older |
| `<` | `<3.0.0` | Version must be strictly older than 3.0.0 |
| `=` | `=1.24.2` | Version must match exactly |
| *(empty)* | `""` | Any version — just check if installed |

---

## Development

### Prerequisites

- Go 1.24+

### Commands

```bash
make build      # Compile binary to bin/
make test       # Run all tests
make test-race  # Run with race detector
make coverage   # Generate HTML coverage report
make lint       # Run golangci-lint
make clean      # Remove build artifacts
```

### Project Structure

```
env-validator-mcp/
├── cmd/server/          # Entry point (main.go)
├── internal/
│   ├── checker/         # Tool, env, and file checkers (interfaces + implementations)
│   ├── report/          # Result and Report types
│   └── semver/          # Version extraction and constraint checking
├── mcp/                 # MCP server setup and tool handlers
├── Makefile
└── go.mod
```

### Architecture: Dependency Injection for TDD

All external effects are hidden behind interfaces:

```
Commander   → exec.Command (real) / MockCommander (tests)
EnvReader   → os.LookupEnv (real) / MockEnvReader (tests)
FileSystem  → os.Stat/ReadFile (real) / MockFileSystem (tests)
```

This allows every checker to be tested in complete isolation — no spawned processes, no real environment, no filesystem I/O.

---

## License

MIT
