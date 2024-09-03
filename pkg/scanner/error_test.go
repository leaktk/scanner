package scanner

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestL(t *testing.T) {
	e := LeakTKError{
		Fatal:   false,
		Code:    2,
		Message: "test error message",
	}
	t.Run("non fatal output", func(t *testing.T) {
		assert.Equal(t, "error occured, code 2 (ScanError): test error message", e.String())
	})

	e.Fatal = true
	t.Run("fatal output", func(t *testing.T) {
		assert.Equal(t, "fatal error occured, code 2 (ScanError): test error message", e.String())
	})
}
