package config

import (
	"fmt"
	"os"
	"testing"

	"github.com/adrg/xdg"
	"github.com/stretchr/testify/assert"
)

func TestPartialLoadConfigFromFile(t *testing.T) {
	cfg, err := LoadConfigFromFile("../../testdata/partial-config.toml")

	if err != nil {
		t.Errorf("Load returned an error %v", err)
	}

	if cfg == nil {
		t.Error("Got a nil config")
	}

	// Check values
	tests := []struct {
		expected string
		actual   string
	}{
		{
			expected: "7.6.1",
			actual:   cfg.Scanner.Gitleaks.Version,
		},
		{
			expected: "/tmp/leaktk/scanner",
			actual:   cfg.Scanner.Workdir,
		},
		{
			expected: "43200",
			actual:   fmt.Sprint(cfg.Scanner.Patterns.RefreshInterval),
		},
		{
			expected: "https://example.com/leaktk/patterns/main/target",
			actual:   cfg.Scanner.Patterns.Server.URL,
		},
		{
			expected: "",
			actual:   cfg.Scanner.Patterns.Server.AuthToken,
		},
		{
			expected: "INFO",
			actual:   cfg.Logger.Level,
		},
	}

	for _, test := range tests {
		assert.Equal(t, test.expected, test.actual)
	}
}

func TestLoadConfigFromFileWithInvalidLogLevel(t *testing.T) {
	cfg, err := LoadConfigFromFile("../../testdata/invalid-log-level.toml")
	assert.Error(t, err, "Expected error from invalid config")
	assert.Nil(t, cfg, "Expected invalid config to be nil")
}

func TestLocateAndLoadConfig(t *testing.T) {
	// Set the env var here to prove the provided path overrides it
	xdg.ConfigHome = "../../testdata/locator-test"
	os.Setenv("LEAKTK_CONFIG", "../../testdata/locator-test/leaktk/config.2.toml")

	// Confirm load from file works
	cfg, err := LocateAndLoadConfig("../../testdata/locator-test/leaktk/config.1.toml")
	assert.Nil(t, err)
	assert.Equal(t, "test-1", cfg.Scanner.Gitleaks.Version)

	// Confirm load from the LEAKTK_CONFIG env var works
	cfg, err = LocateAndLoadConfig("")
	assert.Nil(t, err)
	assert.Equal(t, "test-2", cfg.Scanner.Gitleaks.Version)

	// Confirm load from the LEAKTK_CONFIG env var works
	os.Unsetenv("LEAKTK_CONFIG")
	cfg, err = LocateAndLoadConfig("")
	assert.Nil(t, err)
	assert.Equal(t, "test-3", cfg.Scanner.Gitleaks.Version)

}
