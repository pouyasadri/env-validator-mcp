package checker_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pouyasadri/env-validator-mcp/internal/checker"
	"github.com/pouyasadri/env-validator-mcp/internal/report"
)

func TestCheckConfigFile_Exists(t *testing.T) {
	fs := NewMockFileSystem()
	fs.AddFile("/project/.env", "DATABASE_URL=postgres://localhost/mydb")

	cc := checker.NewConfigChecker(fs)
	result, err := cc.Check("/project/.env", checker.FileSpec{})

	require.NoError(t, err)
	assert.Equal(t, report.StatusOK, result.Status)
	assert.Equal(t, "/project/.env", result.Name)
	assert.Contains(t, result.Message, "exists")
}

func TestCheckConfigFile_Missing(t *testing.T) {
	fs := NewMockFileSystem()
	cc := checker.NewConfigChecker(fs)

	result, err := cc.Check("/project/docker-compose.yml", checker.FileSpec{})

	require.NoError(t, err)
	assert.Equal(t, report.StatusMissing, result.Status)
	assert.Contains(t, result.Message, "not found")
}

func TestCheckConfigFile_Directory_Exists(t *testing.T) {
	fs := NewMockFileSystem()
	fs.AddDir("/project/.git")
	cc := checker.NewConfigChecker(fs)

	result, err := cc.Check("/project/.git", checker.FileSpec{})
	require.NoError(t, err)
	assert.Equal(t, report.StatusOK, result.Status)
}

func TestCheckConfigFile_ValidJSON(t *testing.T) {
	fs := NewMockFileSystem()
	fs.AddFile("/project/package.json", `{"name":"my-app","version":"1.0.0"}`)
	cc := checker.NewConfigChecker(fs)

	result, err := cc.Check("/project/package.json", checker.FileSpec{ValidateJSON: true})
	require.NoError(t, err)
	assert.Equal(t, report.StatusOK, result.Status)
}

func TestCheckConfigFile_InvalidJSON(t *testing.T) {
	fs := NewMockFileSystem()
	fs.AddFile("/project/package.json", `{invalid json}`)
	cc := checker.NewConfigChecker(fs)

	result, err := cc.Check("/project/package.json", checker.FileSpec{ValidateJSON: true})
	require.NoError(t, err)
	assert.Equal(t, report.StatusMisconfigured, result.Status)
	assert.Contains(t, result.Message, "invalid JSON")
}

func TestCheckConfigFile_RequiredKeys_Present(t *testing.T) {
	fs := NewMockFileSystem()
	fs.AddFile("/project/package.json", `{"name":"my-app","version":"1.0.0","scripts":{}}`)
	cc := checker.NewConfigChecker(fs)

	result, err := cc.Check("/project/package.json", checker.FileSpec{
		ValidateJSON: true,
		RequiredKeys: []string{"name", "version"},
	})
	require.NoError(t, err)
	assert.Equal(t, report.StatusOK, result.Status)
}

func TestCheckConfigFile_RequiredKeys_Missing(t *testing.T) {
	fs := NewMockFileSystem()
	fs.AddFile("/project/package.json", `{"name":"my-app"}`)
	cc := checker.NewConfigChecker(fs)

	result, err := cc.Check("/project/package.json", checker.FileSpec{
		ValidateJSON: true,
		RequiredKeys: []string{"name", "version", "main"},
	})
	require.NoError(t, err)
	assert.Equal(t, report.StatusMisconfigured, result.Status)
	assert.Contains(t, result.Message, "version")
	assert.Contains(t, result.Message, "main")
}

func TestCheckConfigFiles_Multiple(t *testing.T) {
	fs := NewMockFileSystem()
	fs.AddFile("/project/go.mod", "module example.com/app\n\ngo 1.24")
	fs.AddFile("/project/.env", "PORT=8080")

	cc := checker.NewConfigChecker(fs)
	specs := map[string]checker.FileSpec{
		"/project/go.mod":             {},
		"/project/.env":               {},
		"/project/docker-compose.yml": {},
	}

	results := cc.CheckMultiple(specs)
	require.Len(t, results, 3)

	statusByName := make(map[string]report.Status)
	for _, r := range results {
		statusByName[r.Name] = r.Status
	}
	assert.Equal(t, report.StatusOK, statusByName["/project/go.mod"])
	assert.Equal(t, report.StatusOK, statusByName["/project/.env"])
	assert.Equal(t, report.StatusMissing, statusByName["/project/docker-compose.yml"])
}
