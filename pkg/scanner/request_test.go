package scanner

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

const validGitRepoRequest = `
{
  "id": "foobar",
  "kind": "GitRepo",
  "resource": "https://github.com/leaktk/fake-leaks.git",
  "options": {
    "depth": 256,
    "since": "2000-01-01"
  }
}
`

const invalidGitRepoRequest = `
{
  "id": "foobar",
  "kind": "GitRepo",
  "options": {
    "depth": true,
    "since": "2000-01-01"
  }
}
`

func TestGitRepoRequest(t *testing.T) {
	var validRequest Request
	err := json.Unmarshal([]byte(validGitRepoRequest), &validRequest)
	assert.Nil(t, err)

	assert.Equal(t, validRequest.ID, "foobar")
	assert.Equal(t, validRequest.Kind, "GitRepo")

	options := validRequest.GitRepoOptions()
	assert.NotNil(t, options)

	assert.Equal(t, options.Depth, 256)
	assert.Equal(t, options.Since, "2000-01-01")
	assert.Equal(t, options.Branch, "")

	var invalidRequest Request
	err = json.Unmarshal([]byte(invalidGitRepoRequest), &invalidRequest)
	assert.NotNil(t, err)
}
