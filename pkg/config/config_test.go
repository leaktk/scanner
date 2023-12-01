package config

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

// TestFullLoadConfigFromFile tests that the config can be completely set
// by a config file
func TestFullLoadConfigFromFile(t *testing.T) {
	config, err := LoadConfigFromFile("../../testdata/full-config.toml")

	if err != nil {
		t.Errorf("Load returned an error %v", err)
	}

	if config == nil {
		t.Error("Got a nil config")
	}

	// Check values
	tests := []struct {
		expected string
		actual   string
	}{
		{
			expected: "set-by-config",
			actual:   config.Scanner.Gitleaks.Version,
		},
		{
			expected: "/tmp/leaktk/scanner",
			actual:   config.Scanner.Workdir,
		},
		{
			expected: "10",
			actual:   fmt.Sprint(config.Scanner.Patterns.RefreshInterval),
		},
		{
			expected: "https://example.com/leaktk/patterns/main/target",
			actual:   config.Scanner.Patterns.Server.URL,
		},
		{
			expected: "placeholder_auth_token",
			actual:   config.Scanner.Patterns.Server.AuthToken,
		},
		{
			expected: "DEBUG",
			actual:   config.Logger.Level,
		},
	}

	for _, test := range tests {
		assert.Equal(t, test.expected, test.actual)
	}
}

func TestPartialLoadConfigFromFile(t *testing.T) {
	config, err := LoadConfigFromFile("../../testdata/partial-config.toml")

	if err != nil {
		t.Errorf("Load returned an error %v", err)
	}

	if config == nil {
		t.Error("Got a nil config")
	}

	// Check values
	tests := []struct {
		expected string
		actual   string
	}{
		{
			expected: "7.6.1",
			actual:   config.Scanner.Gitleaks.Version,
		},
		{
			expected: "/tmp/leaktk/scanner",
			actual:   config.Scanner.Workdir,
		},
		{
			expected: "43200",
			actual:   fmt.Sprint(config.Scanner.Patterns.RefreshInterval),
		},
		{
			expected: "https://example.com/leaktk/patterns/main/target",
			actual:   config.Scanner.Patterns.Server.URL,
		},
		{
			expected: "",
			actual:   config.Scanner.Patterns.Server.AuthToken,
		},
		{
			expected: "INFO",
			actual:   config.Logger.Level,
		},
	}

	for _, test := range tests {
		assert.Equal(t, test.expected, test.actual)
	}
}

func TestLoadConfigFromFileWithInvalidLogLevel(t *testing.T) {
	config, err := LoadConfigFromFile("../../testdata/invalid-log-level.toml")

	if err == nil {
		t.Error("An error was not returned")
	}

	if config != nil {
		t.Errorf("Config was returned when it shouldn't be: %v", config)
	}
}
