package resource

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFiles(t *testing.T) {
	// Set up dir structure
	tmpDir := t.TempDir()
	err := os.MkdirAll(filepath.Join(tmpDir, "foo"), 0700)
	assert.NoError(t, err)

	// Write test file
	testFileData := []byte("Hello, world!\n")
	testFilePath := filepath.Join(tmpDir, "foo", "test-file")
	err = os.WriteFile(testFilePath, testFileData, 0600)
	assert.NoError(t, err)

	// Create testing object
	files := NewFiles(tmpDir, &FilesOptions{})

	t.Run("ReadFile", func(t *testing.T) {
		readData, err := files.ReadFile(filepath.Join("foo", "test-file"))
		assert.NoError(t, err)
		assert.Equal(t, string(testFileData), string(readData))
	})

	t.Run("Objects", func(t *testing.T) {
		_ = files.Objects(func(obj Object) error {
			data, err := io.ReadAll(obj.Content)
			assert.NoError(t, err)
			assert.Equal(t, obj.Path, filepath.Join("foo", "test-file"))
			assert.Equal(t, string(testFileData), string(data))
			return nil
		})
	})
}
