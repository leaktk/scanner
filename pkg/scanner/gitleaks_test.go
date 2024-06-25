package scanner

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/leaktk/scanner/pkg/config"
	"github.com/leaktk/scanner/pkg/logger"
	"github.com/leaktk/scanner/pkg/resource"
	"github.com/stretchr/testify/assert"
)

const mockGitleaksTestConfig = `
[[rules]]
id = "test-rule"
description = "A rule for finding some value in the fake-leaks repo"
regex = '''(?i)(fake)'''
secretGroup = 1
`

func TestGitleaksScan(t *testing.T) {
	err := logger.SetLoggerLevel("CRITICAL")
	assert.NoError(t, err)

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "gitleaks.toml")
	err = os.WriteFile(configPath, []byte(mockGitleaksTestConfig), 0644)
	assert.NoError(t, err)
	cfg := config.DefaultConfig()
	cfg.Scanner.Patterns.Gitleaks.ConfigPath = configPath

	// Configured patterns
	patterns := NewPatterns(
		&cfg.Scanner.Patterns,
		// There shouldn't be a need for a HTTP client since the patterns are fresh
		nil,
	)

	t.Run("Success", func(t *testing.T) {
		assert.NoError(t, err)

		gitRepo := resource.NewGitRepo("https://github.com/leaktk/fake-leaks.git", &resource.GitRepoOptions{
			Branch: "main",
			Depth:  1000,
		})

		err = gitRepo.Clone(filepath.Join(tempDir, "repo"))
		assert.NoError(t, err)

		results, err := NewGitleaks(patterns).Scan(gitRepo)
		assert.NoError(t, err)
		assert.Greater(t, len(results), 0)
		assert.Equal(t, results[0].Rule.ID, "test-rule")
		assert.Equal(t, results[0].Secret, "fake")
	})
}
