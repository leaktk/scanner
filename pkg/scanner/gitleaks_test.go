package scanner

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/leaktk/leaktk/pkg/config"
	"github.com/leaktk/leaktk/pkg/logger"
	"github.com/leaktk/leaktk/pkg/resource"
)

const mockGitleaksTestConfig = `
[[rules]]
id = "test-rule"
description = "A rule for finding some value in the fake-leaks repo"
regex = '''(?i)(fake)'''
secretGroup = 1
`

const mockInvalidGitleaksTestConfig = `
[[rules]]
id = "test-rule"
description = "A rule for finding some value in the fake-leaks repo"
regex = '''(?!@#!i)(fake)'''
secretGroup = 1
`

func TestGitleaksScan(t *testing.T) {
	err := logger.SetLoggerLevel("CRITICAL")
	assert.NoError(t, err)

	t.Run("Success", func(t *testing.T) {
		assert.NoError(t, err)
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "gitleaks.toml")
		err = os.WriteFile(configPath, []byte(mockGitleaksTestConfig), 0600)
		assert.NoError(t, err)
		cfg := config.DefaultConfig()
		cfg.Scanner.Patterns.Gitleaks.ConfigPath = configPath

		// Configured patterns
		patterns := NewPatterns(
			&cfg.Scanner.Patterns,
			// There shouldn't be a need for a HTTP client since the patterns are fresh
			nil,
		)

		gitRepo := resource.NewGitRepo("https://github.com/leaktk/fake-leaks.git", &resource.GitRepoOptions{
			Branch: "main",
			Depth:  1000,
		})

		err = gitRepo.Clone(filepath.Join(tempDir, "clone"))
		assert.NoError(t, err)

		results, err := NewGitleaks(1, patterns).Scan(gitRepo)
		assert.NoError(t, err)
		assert.Greater(t, len(results), 0)
		// This should at least be defined on git responses
		assert.Contains(t, results[0].Notes["gitleaks_fingerprint"], ":test-rule:")
		assert.Equal(t, results[0].Rule.ID, "test-rule")
		assert.Equal(t, strings.ToLower(results[0].Secret), "fake")
	})

	t.Run("Invalid Config", func(t *testing.T) {
		assert.NoError(t, err)
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "gitleaks.toml")
		err = os.WriteFile(configPath, []byte(mockInvalidGitleaksTestConfig), 0600)
		assert.NoError(t, err)
		cfg := config.DefaultConfig()
		cfg.Scanner.Patterns.Gitleaks.ConfigPath = configPath

		// Configured patterns
		patterns := NewPatterns(
			&cfg.Scanner.Patterns,
			// There shouldn't be a need for a HTTP client since the patterns are fresh
			nil,
		)

		gitRepo := resource.NewGitRepo("https://github.com/leaktk/fake-leaks.git", &resource.GitRepoOptions{
			Branch: "main",
			Depth:  1000,
		})

		err = gitRepo.Clone(filepath.Join(tempDir, "clone"))
		assert.NoError(t, err)

		results, err := NewGitleaks(1, patterns).Scan(gitRepo)
		assert.Error(t, err)
		assert.Equal(t, len(results), 0)
	})
}
