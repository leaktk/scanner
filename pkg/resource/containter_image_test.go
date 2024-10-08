package resource

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContainerImage(t *testing.T) {
	t.Run("Clone", func(t *testing.T) {
		tempDir := t.TempDir()

		image := NewContainerImage("quay.io/leaktk/fake-leaks:v1.0.1", &ContainerImageOptions{})

		err := image.Clone(tempDir)
		assert.NoError(t, err)
		contact := image.Contact()
		assert.Equal(t, "Fake Leaks", contact.Name)
		assert.Equal(t, "fake-leaks@leaktk.org", contact.Email)
	})

}
