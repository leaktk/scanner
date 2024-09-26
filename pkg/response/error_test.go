package response

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestL(t *testing.T) {
	e := LeakTKError{
		Fatal:   false,
		Code:    2,
		Message: "test error message",
	}
	t.Run("non fatal output", func(t *testing.T) {
		assert.Equal(t, "error occurred, code 2 (ScanError): test error message", e.String())
	})

	e.Fatal = true
	t.Run("fatal output", func(t *testing.T) {
		assert.Equal(t, "fatal error occurred, code 2 (ScanError): test error message", e.String())
	})
}
