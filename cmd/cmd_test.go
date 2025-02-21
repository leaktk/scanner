package cmd

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/leaktk/scanner/pkg/fs"
)

func TestScanCommandToRequest(t *testing.T) {
	cmd := scanCommand()

	// Resource must be set
	request, err := scanCommandToRequest(cmd)
	assert.Nil(t, request)
	assert.Error(t, err)
	assert.Equal(t, err.Error(), "missing required field: field=\"resource\"")

	// Setting resource for the rest of the tests
	_ = cmd.Flags().Set("resource", "https://github.com/leaktk/fake-leaks.git")
	request, err = scanCommandToRequest(cmd)
	assert.NoError(t, err)
	assert.NotNil(t, request)

	// ID should default to a random id
	assert.Equal(t, 11, len(request.ID))
	// Kind should default to GitRepo
	assert.Equal(t, request.Resource.Kind(), "GitRepo")
	assert.Equal(t, request.Resource.String(), "https://github.com/leaktk/fake-leaks.git")

	// If resource starts with @ and the thing is a valid path, resource will be loaded from there
	tmpDir := t.TempDir()
	data_path, err := fs.CleanJoin(tmpDir, "data.json")
	assert.NoError(t, err)
	err = os.WriteFile(data_path, []byte("{\"some\": \"data\"}"), 0600)
	assert.NoError(t, err)

	_ = cmd.Flags().Set("resource", "@"+data_path)
	_ = cmd.Flags().Set("kind", "JSONData")
	request, err = scanCommandToRequest(cmd)
	assert.NoError(t, err)
	assert.Equal(t, request.Resource.Kind(), "JSONData")
	assert.Equal(t, request.Resource.String(), "{\"some\": \"data\"}")

	// If resource starts with @ and the thing is an invalid path, raise an error
	_ = cmd.Flags().Set("resource", "@"+data_path+".invalid")
	request, err = scanCommandToRequest(cmd)
	assert.Error(t, err)
	assert.Nil(t, request)
	assert.Equal(t, err.Error(), fmt.Sprintf("resource path does not exist: path=%q", data_path+".invalid"))
}
