package resource

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJSONDataReadFile(t *testing.T) {
	data := `{"foo":"bar", "baz": ["bop", true, 1, 2.3, null, {"hello": "there"}]}`
	jsonData := NewJSONData(data, &JSONDataOptions{})
	err := jsonData.Clone(t.TempDir())
	assert.NoError(t, err)

	t.Run("Success", func(t *testing.T) {
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
				path:  "baz/0",
				value: "bop",
				err:   false,
			},
			{
				path:  "//////baz/0",
				value: "bop",
				err:   false,
			},
			{
				path:  "/baz/1",
				value: "true",
				err:   false,
			},
			{
				path:  "/baz/2",
				value: "1",
				err:   false,
			},
			{
				path:  "/baz/3",
				value: "2.3",
				err:   false,
			},
			{
				path:  "/baz/4",
				value: "",
				err:   false,
			}, {
				path:  "/baz/5/hello",
				value: "there",
				err:   false,
			},
			{
				path: "baz",
				err:  true,
			},
			{
				path: "baz/10",
				err:  true,
			},
			{
				path: "baz/fish",
				err:  true,
			},
			{
				path: "cat",
				err:  true,
			},
		}

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
}