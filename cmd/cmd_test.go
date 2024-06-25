package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScanCommandToRequest(t *testing.T) {
	cmd := scanCommand()

	// Resource must be set
	request, err := scanCommandToRequest(cmd)
	assert.Nil(t, request)
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "missing required field: resource")

	// Setting resource for the rest of the tests
	cmd.Flags().Set("resource", "https://github.com/leaktk/fake-leaks.git")
	request, err = scanCommandToRequest(cmd)
	assert.Nil(t, err)
	assert.NotNil(t, request)

	// Id should default to a random UUID
	assert.Equal(t, 64, len(request.ID))
	// Kind should default to GitRepo
	assert.Equal(t, request.Resource.Kind(), "GitRepo")
	assert.Equal(t, request.Resource.String(), "https://github.com/leaktk/fake-leaks.git")
}
