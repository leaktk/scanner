package resource

import (
	"fmt"
	"io"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJSONData(t *testing.T) {
	data := `{"foo":"bar", "baz": ["bop", true, 1, 2.3, null, {"hello": "there"}]}`
	jsonData := NewJSONData(data, &JSONDataOptions{})
	err := jsonData.Clone(t.TempDir())
	assert.NoError(t, err)

	tests := []struct {
		path  string
		value string
		err   bool
	}{
		{
			path:  "foo",
			value: "bar",
			err:   false,
		},
		{
			path:  filepath.Join("baz", "0"),
			value: "bop",
			err:   false,
		},
		{
			path:  filepath.Join("baz", "0"),
			value: "bop",
			err:   false,
		},
		{
			path:  filepath.Join("baz", "1"),
			value: "true",
			err:   false,
		},
		{
			path:  filepath.Join("baz", "2"),
			value: "1",
			err:   false,
		},
		{
			path:  filepath.Join("baz", "3"),
			value: "2.3",
			err:   false,
		},
		{
			path:  filepath.Join("baz", "4"),
			value: "",
			err:   false,
		}, {
			path:  filepath.Join("baz", "5", "hello"),
			value: "there",
			err:   false,
		},
		{
			path: "baz",
			err:  true,
		},
		{
			path: filepath.Join("baz", "10"),
			err:  true,
		},
		{
			path: filepath.Join("baz", "fish"),
			err:  true,
		},
		{
			path: "cat",
			err:  true,
		},
	}

	t.Run("ReadFile", func(t *testing.T) {
		for _, test := range tests {
			value, err := jsonData.ReadFile(test.path)

			if !test.err {
				assert.NoError(t, err, "path: %s", test.path)
				assert.Equal(t, test.value, string(value), "path: %s", test.path)
			} else {
				assert.Error(t, err, "path: %s", test.path)
			}
		}
	})

	t.Run("Walk", func(t *testing.T) {
		toCheck := map[string]string{}

		// Build out a dict to check
		for _, test := range tests {
			// Ignore the error tests for this one since walk should only return
			// valid paths
			if !test.err {
				// Walk only returns relative paths
				if filepath.IsAbs(test.path) {
					relPath, err := filepath.Rel(string(filepath.Separator), test.path)
					assert.NoError(t, err)
					toCheck[relPath] = test.value
				} else {
					toCheck[test.path] = test.value
				}
			}
		}

		// Walk the tests to make sure the items are present
		_ = jsonData.Walk(func(path string, reader io.Reader) error {
			data, err := io.ReadAll(reader)
			assert.NoError(t, err)
			expected, exists := toCheck[path]

			if exists {
				assert.Equal(t, expected, string(data))
				delete(toCheck, path)
			}

			return nil
		})

		// Make sure all items have been checked
		assert.Equal(t, len(toCheck), 0, fmt.Sprintf("remaining values: %v", toCheck))
	})
}
