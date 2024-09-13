package resource

import (
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

func TestContainerImage(t *testing.T) {
	t.Run("SanitizePath", func(t *testing.T) {
		tmpDir := t.TempDir()
		err := os.MkdirAll(filepath.Join(tmpDir, "foo"), 0700)
		assert.NoError(t, err)

		testPathFail := "../../hello/world"
		_, err = sanitizePath(tmpDir, testPathFail)
		assert.Error(t, err)

		testPathPass := "hello/world..zip"
		_, err = sanitizePath(tmpDir, testPathPass)
		assert.NoError(t, err)
	})

	t.Run("Clone", func(t *testing.T) {
		tempDir := t.TempDir()

		image := NewContainerImage("quay.io/wizzy/fake-leaks:v1.0.2", &ContainerImageOptions{})

		err := image.Clone(tempDir)
		assert.NoError(t, err)
		name, email := image.Contact()
		assert.Equal(t, "Josh Maint", name)
		assert.Equal(t, "wizzy-maint@wizzy.com", email)
	})

}
