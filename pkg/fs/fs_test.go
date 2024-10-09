package fs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileExists(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "file")

	t.Run("DirIsNotAFile", func(t *testing.T) {
		assert.False(t, FileExists(tmpDir))
	})

	t.Run("FileExists", func(t *testing.T) {
		err := os.WriteFile(tmpFile, []byte{}, 0600)
		assert.NoError(t, err)
		assert.True(t, FileExists(tmpFile))
	})

	t.Run("FileDoesntExist", func(t *testing.T) {
		noFile := filepath.Join(tmpFile, "foo/bar/baz")
		assert.False(t, FileExists(noFile))
	})
}

func TestPathExists(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "file")

	t.Run("DirExists", func(t *testing.T) {
		assert.True(t, PathExists(tmpDir))
	})

	t.Run("FileExists", func(t *testing.T) {
		err := os.WriteFile(tmpFile, []byte{}, 0600)
		assert.NoError(t, err)
		assert.True(t, PathExists(tmpFile))
	})

	t.Run("DirDoesntExist", func(t *testing.T) {
		noDir := filepath.Join(tmpDir, "foo/bar/baz")
		assert.False(t, PathExists(noDir))
	})

	t.Run("FileDoesntExist", func(t *testing.T) {
		noFile := filepath.Join(tmpFile, "foo/bar/baz")
		assert.False(t, PathExists(noFile))
	})
}

func TestCleanJoin(t *testing.T) {
	t.Run("CleanJoin", func(t *testing.T) {
		tmpDir := t.TempDir()
		err := os.MkdirAll(filepath.Join(tmpDir, "foo"), 0700)
		assert.NoError(t, err)

		testPathFail := "../../hello/world"
		_, err = CleanJoin(tmpDir, testPathFail)
		assert.Error(t, err)

		testPathPass := "hello/world..zip"
		_, err = CleanJoin(tmpDir, testPathPass)
		assert.NoError(t, err)
	})
}
