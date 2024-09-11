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
}
