package id

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestID(t *testing.T) {
	t.Run("NoParamsCreatesRandomID", func(t *testing.T) {
		assert.NotEqual(t, ID(), ID())
	})

	t.Run("IDsAreSameLength", func(t *testing.T) {
		assert.Equal(t, 64, len(ID()))
		assert.Equal(t, 64, len(ID("foo")))
		assert.Equal(t, 64, len(ID("foo", "bar")))
	})

	t.Run("IDsAreHexadecimal", func(t *testing.T) {
		// Run the test a bunch of times on random and parameterized ID calls
		for i := 0; i < 100; i++ {
			assert.Regexp(t, "^[0-9a-f]+$", ID())
			assert.Regexp(t, "^[0-9a-f]+$", ID(strconv.Itoa(i), ID()))
		}
	})
}
