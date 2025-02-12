package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPartialLoadConfigFromFile(t *testing.T) {
	os.Setenv("LEAKTK_PATTERN_SERVER_AUTH_TOKEN", "x")
	os.Unsetenv("LEAKTK_PATTERN_SERVER_URL")
	cfg, err := LoadConfigFromFile("../../testdata/partial-config.toml")

	if err != nil {
		// If there are config issues fail fast
		assert.FailNowf(t, "Failed to load config file", "Load returned an error %s", err)
	}

	// Check values
	tests := []struct {
		expected any
		actual   any
	}{
		{
			expected: "8.18.2",
			actual:   cfg.Scanner.Patterns.Gitleaks.Version,
		},
		{
			expected: "/tmp/leaktk/scanner",
			actual:   cfg.Scanner.Workdir,
		},
		{
			expected: uint32(43200),
			actual:   cfg.Scanner.Patterns.RefreshAfter,
		},
		{
			expected: "https://example.com/leaktk/patterns/main/target",
			actual:   cfg.Scanner.Patterns.Server.URL,
		},
		{
			expected: "x",
			actual:   cfg.Scanner.Patterns.Server.AuthToken,
		},
		{
			expected: "INFO",
			actual:   cfg.Logger.Level,
		},
		{
			expected: uint16(0),
			actual:   cfg.Scanner.MaxScanDepth,
		},
	}

	for _, test := range tests {
		assert.Equal(t, test.expected, test.actual)
	}
}

func TestLocateAndLoadConfig(t *testing.T) {
	// Set the env var here to prove the provided path overrides it
	localConfigDir = "../../testdata/locator-test/leaktk"
	os.Setenv("LEAKTK_CONFIG_PATH", "../../testdata/locator-test/leaktk/config.2.toml")

	// Confirm load from file works
	cfg, err := LocateAndLoadConfig("../../testdata/locator-test/leaktk/config.1.toml")
	assert.Nil(t, err)
	assert.Equal(t, "test-1", cfg.Scanner.Patterns.Gitleaks.Version)

	// Confirm load from the LEAKTK_CONFIG_PATH env var works
	cfg, err = LocateAndLoadConfig("")
	assert.Nil(t, err)
	assert.Equal(t, "test-2", cfg.Scanner.Patterns.Gitleaks.Version)

	// Confirm load from the LEAKTK_CONFIG_PATH env var works
	os.Unsetenv("LEAKTK_CONFIG_PATH")
	cfg, err = LocateAndLoadConfig("")
	assert.Nil(t, err)
	assert.Equal(t, "test-3", cfg.Scanner.Patterns.Gitleaks.Version)

}
