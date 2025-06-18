package scanner

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/leaktk/leaktk/pkg/config"
	httpclient "github.com/leaktk/leaktk/pkg/http"
)

const mockConfig = `
[allowlist]
paths = ['''testdata''']

[[rules]]
id = "test-rule"
description = "test-rule"
regex = '''test-rule'''
`

func TestPatternsFetchGitleaksConfig(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Equal(t, "/patterns/gitleaks/x.y.z", r.URL.Path)
			w.WriteHeader(http.StatusOK)
			_, err := io.WriteString(w, mockConfig)
			assert.NoError(t, err)
		}))
		ts.Start()
		defer ts.Close()

		cfg := config.DefaultConfig()
		cfg.Scanner.Patterns.Server.URL = ts.URL
		cfg.Scanner.Patterns.Gitleaks.Version = "x.y.z"

		client := httpclient.NewClient()
		p := NewPatterns(&cfg.Scanner.Patterns, client)

		rawConfig, err := p.fetchGitleaksConfig()
		assert.NoError(t, err)
		assert.Contains(t, rawConfig, "test-rule")
	})

	t.Run("InvalidURL", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.Scanner.Patterns.Server.URL = "invalid-url"
		cfg.Scanner.Patterns.Gitleaks.Version = "x.y.z"

		client := httpclient.NewClient()
		p := NewPatterns(&cfg.Scanner.Patterns, client)

		_, err := p.fetchGitleaksConfig()
		assert.Error(t, err)
	})

	t.Run("HTTPError", func(t *testing.T) {
		ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		ts.Start()
		defer ts.Close()

		cfg := config.DefaultConfig()
		cfg.Scanner.Patterns.Server.URL = ts.URL
		cfg.Scanner.Patterns.Gitleaks.Version = "x.y.z"

		client := httpclient.NewClient()
		p := NewPatterns(&cfg.Scanner.Patterns, client)

		_, err := p.fetchGitleaksConfig()
		assert.Error(t, err)
	})

	t.Run("WithAuthToken", func(t *testing.T) {
		ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Equal(t, "/patterns/gitleaks/x.y.z", r.URL.Path)
			assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
			w.WriteHeader(http.StatusOK)
			_, err := io.WriteString(w, mockConfig)
			assert.NoError(t, err)
		}))
		ts.Start()
		defer ts.Close()

		cfg := config.DefaultConfig()
		cfg.Scanner.Patterns.Server.URL = ts.URL
		cfg.Scanner.Patterns.Server.AuthToken = "test-token"
		cfg.Scanner.Patterns.Gitleaks.Version = "x.y.z"

		client := httpclient.NewClient()
		p := NewPatterns(&cfg.Scanner.Patterns, client)

		rawConfig, err := p.fetchGitleaksConfig()
		assert.NoError(t, err)
		assert.Contains(t, rawConfig, "test-rule")
	})
}

func TestGitleaksConfigModTimeExceeds(t *testing.T) {
	t.Run("FileExistsAndOlderThanLimit", func(t *testing.T) {
		tempDir := t.TempDir()

		tempFilePath := filepath.Join(tempDir, "gitleaks.toml")
		err := os.WriteFile(tempFilePath, []byte{}, 0600)
		assert.NoError(t, err)

		// Set the file's modification time to 10 seconds ago
		err = os.Chtimes(tempFilePath, time.Now().Add(-10*time.Second), time.Now().Add(-10*time.Second))
		assert.NoError(t, err)

		// Create a Patterns instance with the temporary file path
		cfg := config.DefaultConfig()
		cfg.Scanner.Patterns.Gitleaks.ConfigPath = tempFilePath

		patterns := &Patterns{
			config: &cfg.Scanner.Patterns,
		}

		// Test with a modTimeLimit of 5 seconds
		assert.True(t, patterns.gitleaksConfigModTimeExceeds(5))

		// Test with a modTimeLimit of 15 seconds
		assert.False(t, patterns.gitleaksConfigModTimeExceeds(15))
	})

	t.Run("FileDoesNotExist", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.Scanner.Patterns.Gitleaks.ConfigPath = "/path/to/nonexistent/file.toml"

		// Create a Patterns instance with a non-existent file path
		patterns := &Patterns{
			config: &cfg.Scanner.Patterns,
		}

		// Test with any modTimeLimit
		assert.True(t, patterns.gitleaksConfigModTimeExceeds(5))
		assert.True(t, patterns.gitleaksConfigModTimeExceeds(15))
	})

	t.Run("FileExistsButErrorOnStat", func(t *testing.T) {
		// Create a Patterns instance with a file path that causes an error on Stat
		cfg := config.DefaultConfig()
		cfg.Scanner.Patterns.Gitleaks.ConfigPath = "/dev/zero"

		patterns := &Patterns{
			config: &cfg.Scanner.Patterns,
		}

		// Test with any modTimeLimit
		assert.True(t, patterns.gitleaksConfigModTimeExceeds(5))
		assert.True(t, patterns.gitleaksConfigModTimeExceeds(15))
	})
}
