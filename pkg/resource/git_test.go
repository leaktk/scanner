package resource

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGit(t *testing.T) {
	t.Run("BranchNotProvied", func(t *testing.T) {
		tempDir := t.TempDir()

		gitRepo := NewGitRepo("https://github.com/leaktk/fake-leaks.git", &GitRepoOptions{
			Depth: 1,
		})

		err := gitRepo.Clone(tempDir)
		assert.NoError(t, err)
		assert.Greater(t, len(gitRepo.Refs()), 1)
	})

	t.Run("BranchProvied", func(t *testing.T) {
		tempDir := t.TempDir()

		gitRepo := NewGitRepo("https://github.com/leaktk/fake-leaks.git", &GitRepoOptions{
			Branch: "main",
			Depth:  1,
		})

		err := gitRepo.Clone(tempDir)
		assert.NoError(t, err)
		assert.Equal(t, len(gitRepo.Refs()), 1)
	})

	t.Run("BranchInvalid", func(t *testing.T) {
		tempDir := t.TempDir()

		gitRepo := NewGitRepo("https://github.com/leaktk/fake-leaks.git", &GitRepoOptions{
			Branch: "invalid-branch-e7d33d3a-6057-432a-863b-0ed844ee1f7b",
			Depth:  1,
		})

		err := gitRepo.Clone(tempDir)
		assert.NoError(t, err)
		// Should increase the scope of the clone to be the same as if no branch
		// was provided
		assert.Greater(t, len(gitRepo.Refs()), 1)
	})

}
