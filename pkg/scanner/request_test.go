package scanner

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/leaktk/leaktk/pkg/proto"
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
	var validRequest proto.Request
	err := json.Unmarshal([]byte(validGitRepoRequest), &validRequest)
	assert.Nil(t, err)

	assert.Equal(t, validRequest.ID, "foobar")
	assert.Equal(t, validRequest.Kind, proto.GitRepoRequestKind)

	var invalidRequest proto.Request
	err = json.Unmarshal([]byte(invalidGitRepoRequest), &invalidRequest)
	assert.NotNil(t, err)
}
