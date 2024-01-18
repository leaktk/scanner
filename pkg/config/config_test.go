package config

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
	assert.Error(t, err, "Expected error from invalid config")
	assert.Nil(t, config, "Expected invalid config to be nil")
}
