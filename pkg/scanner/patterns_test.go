package scanner

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/leaktk/scanner/pkg/config"
)

const mockConfig = `
[allowlist]
paths = ['''testdata''']

[[rules]]
description = "Find the foo"
regex = '''foo'''
`

type mockHttpClient struct {
}

func (c *mockHttpClient) Do(req *http.Request) (*http.Response, error) {
	return &http.Response{
		Body: io.NopCloser(bytes.NewReader([]byte(mockConfig))),
	}, nil
}

func assertPathExists(t *testing.T, path string) {
	_, err := os.Stat(path)
	assert.Nil(t, err, fmt.Sprintf("the path %s does not exist", path))
}

func assertPathNotExists(t *testing.T, path string) {
	_, err := os.Stat(path)
	assert.NotNil(t, err, fmt.Sprintf("the path %s does exists", path))
}

func TestGitLeaksPatterns(t *testing.T) {
	cfg := &config.DefaultConfig().Scanner.Patterns

	tmpDir, err := os.MkdirTemp("", "leaktk-test.")
	assert.Nil(t, err)
	cfg.Gitleaks.ConfigPath = filepath.Clean(filepath.Join(tmpDir, "gitleaks.toml"))

	client := &mockHttpClient{}
	patterns := NewPatterns(cfg, client)

	// Test fetch when autofetch is off
	assertPathNotExists(t, cfg.Gitleaks.ConfigPath)
	cfg.Autofetch = false
	gitleaks, err := patterns.Gitleaks()
	assert.NotNil(t, err)

	// Test fetch missing patterns
	assertPathNotExists(t, cfg.Gitleaks.ConfigPath)
	cfg.Autofetch = true
	gitleaks, err = patterns.Gitleaks()
	assert.Nil(t, err)
	assert.Equal(t, len(gitleaks.Rules), 1)
	assertPathExists(t, cfg.Gitleaks.ConfigPath)

	// TODO:
	// test loading patterns from the in memory cache before a timeout
	// test loading patterns from the file system before a timeout
	// test that patterns are refetched after the timeout

	// Clean up the tmpDir
	assert.Nil(t, os.RemoveAll(tmpDir))
}
