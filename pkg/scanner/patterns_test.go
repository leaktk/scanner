package scanner

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/leaktk/scanner/pkg/config"
	"github.com/stretchr/testify/assert"
)

const mockConfig = `
[allowlist]
paths = ['''testdata''']

[[rules]]
description = "test-rule"
regex = '''test-rule'''
`
const mockAllowlistOnlyConfig = `
[allowlist]
paths = ['''testdata''']
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

		client := &http.Client{}
		p := NewPatterns(&cfg.Scanner.Patterns, client)

		rawConfig, err := p.fetchGitleaksConfig()
		assert.NoError(t, err)
		assert.Contains(t, rawConfig, "test-rule")
	})

	t.Run("InvalidURL", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.Scanner.Patterns.Server.URL = "invalid-url"
		cfg.Scanner.Patterns.Gitleaks.Version = "x.y.z"

		client := &http.Client{}
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

		client := &http.Client{}
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

		client := &http.Client{}
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
		err := os.WriteFile(tempFilePath, []byte{}, 0644)
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

func TestParseGitleaksConfig(t *testing.T) {
	t.Run("ValidConfig", func(t *testing.T) {
		cfg, err := ParseGitleaksConfig(mockConfig)
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, "testdata", cfg.Allowlist.Paths[0].String())
	})

	t.Run("AllowlistOnlyConfig", func(t *testing.T) {
		cfg, err := ParseGitleaksConfig(mockAllowlistOnlyConfig)
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, "testdata", cfg.Allowlist.Paths[0].String())
	})

	t.Run("InvalidConfig", func(t *testing.T) {
		rawConfig := "\ninvalid_key = \"value\"\n"
		_, err := ParseGitleaksConfig(rawConfig)
		assert.Error(t, err)
	})

	t.Run("EmptyConfig", func(t *testing.T) {
		rawConfig := ""
		_, err := ParseGitleaksConfig(rawConfig)
		assert.Error(t, err)
	})
}

func TestPatternsGitleaks(t *testing.T) {
	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/patterns/gitleaks/x.y.z", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		_, err := io.WriteString(w, mockConfig)
		assert.NoError(t, err)
	}))

	ts.Start()
	defer ts.Close()

	tempDir := t.TempDir()
	configFilePath := filepath.Join(tempDir, "gitleaks.toml")

	getPatterns := func(autofetch bool, refreshAfter, expiredAfter, modTime uint32) *Patterns {
		err := os.WriteFile(configFilePath, []byte(mockConfig), 0644)
		assert.NoError(t, err)

		// Set the file's modification time to 15 seconds ago
		err = os.Chtimes(configFilePath, time.Now().Add(time.Duration(-modTime)*time.Second), time.Now().Add(time.Duration(-modTime)*time.Second))
		assert.NoError(t, err)

		cfg := config.DefaultConfig()
		cfg.Scanner.Patterns.Server.URL = ts.URL
		cfg.Scanner.Patterns.Autofetch = autofetch
		cfg.Scanner.Patterns.RefreshAfter = refreshAfter
		cfg.Scanner.Patterns.ExpiredAfter = expiredAfter
		cfg.Scanner.Patterns.Gitleaks.ConfigPath = configFilePath
		cfg.Scanner.Patterns.Gitleaks.Version = "x.y.z"

		return NewPatterns(&cfg.Scanner.Patterns, &http.Client{})
	}

	t.Run("AutofetchEnabledAndConfigExpired", func(t *testing.T) {
		cfg, err := getPatterns(true, 5, 10, 15).Gitleaks()
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, "testdata", cfg.Allowlist.Paths[0].String())

		// Verify the config file was updated
		data, err := os.ReadFile(configFilePath)
		assert.NoError(t, err)
		assert.Equal(t, mockConfig, string(data))
	})

	t.Run("AutofetchEnabledAndConfigNotExpired", func(t *testing.T) {
		cfg, err := getPatterns(true, 5, 15, 10).Gitleaks()
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, "testdata", cfg.Allowlist.Paths[0].String())

		// Verify the config file was updated
		data, err := os.ReadFile(configFilePath)
		assert.NoError(t, err)
		assert.Equal(t, mockConfig, string(data))
	})

	t.Run("AutofetchDisabledAndConfigExpired", func(t *testing.T) {
		_, err := getPatterns(false, 5, 10, 15).Gitleaks()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "gitleaks config is expired and autofetch is disabled")
	})

	t.Run("AutofetchDisabledAndConfigNotExpired", func(t *testing.T) {
		cfg, err := getPatterns(false, 5, 15, 10).Gitleaks()
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, "testdata", cfg.Allowlist.Paths[0].String())
	})

	t.Run("ValidateGitleaksConfigHash", func(t *testing.T) {
		patterns := getPatterns(false, 5, 15, 10)
		cfg, err := patterns.Gitleaks()
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.NotNil(t, patterns)
		assert.Equal(t, "32359008caab9646f685745b9888a71d81697a3fdfa5a2f49cc8d8f38649320b", patterns.GitleaksConfigHash())
	})
}
