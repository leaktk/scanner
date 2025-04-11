package resource

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestText(t *testing.T) {
	data := `user_secret="GrWTm5k2jtWy9CeOVbOjlREt-10NmpePFKxv4Fml89YLwn002kF1cy4LQ1cXs9d2PGx37zOUPQk1yViMhhIdHlhw"`
	text := NewText(data, &TextOptions{})
	err := text.Clone(t.TempDir())
	assert.NoError(t, err)

	t.Run("ReadFile", func(t *testing.T) {
		value, err := text.ReadFile("anything")
		assert.Error(t, err)

		value, err = text.ReadFile("")
		assert.NoError(t, err)
		assert.Equal(t, string(value), data)
	})

	t.Run("Walk", func(t *testing.T) {
		_ = text.Walk(func(path string, reader io.Reader) error {
			value, err := io.ReadAll(reader)
			assert.NoError(t, err)
			assert.Equal(t, string(value), text.String())
			return nil
		})
	})
}
