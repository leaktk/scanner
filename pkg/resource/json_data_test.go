package resource

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJSONData(t *testing.T) {
	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error
		var content string

		assert.Equal(t, "GET", r.Method)

		switch r.URL.Path {
		case "/hello":
			content = "hello world"
		case "/hello.json":
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			content = `{"hello": "world"}`
		default:
			content = "Not sure what happened here"
		}

		w.WriteHeader(http.StatusOK)
		_, err = io.WriteString(w, content)

		assert.NoError(t, err)
	}))

	ts.Start()
	defer ts.Close()

	data := `{
			"foo": "bar",
			"baz": ["bop", true, 1, 2.3, null, {"hello": "there"}],
			"url": "` + ts.URL + `/hello",
			"nested": {"url": "` + ts.URL + `/hello.json"},
			"skipped": "https://example.com"
	} `

	brokenURLData := `{
		"invalid": "https://raw.githubusercontent.com/leaktk/fake-leaks/main/this-url-doesnt-exist-8UaehX5b24MzZiaeJ428FK5R"
	}`

	jsonData := NewJSONData(data, &JSONDataOptions{
		// fetch url and anything one level under nested
		FetchURLs: "url:nested/*",
	})

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
		{
			path:  "url",
			value: "hello world",
			err:   false,
		},
		{
			path:  "skipped",
			value: "https://example.com",
			err:   false,
		},

		{
			path:  filepath.Join("nested", "url", "hello"),
			value: "world",
			err:   false,
		},
		{
			path:  "invalid",
			value: "https://raw.githubusercontent.com/leaktk/fake-leaks/main/this-url-doesnt-exist-8UaehX5b24MzZiaeJ428FK5R",
			err:   true,
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
		assert.Equal(t, 0, len(toCheck), fmt.Sprintf("remaining values: %v", toCheck))
	})

	t.Run("CloneBrokenURLWithoutFetchURLs", func(t *testing.T) {
		// Should work
		jsonData := NewJSONData(brokenURLData, &JSONDataOptions{})

		assert.NoError(t, jsonData.Clone(t.TempDir()))
	})

	t.Run("CloneBrokenURLWithFetchURLs", func(t *testing.T) {
		jsonData := NewJSONData(brokenURLData, &JSONDataOptions{
			// Fetch anything that ends with invalid
			FetchURLs: "**/invalid",
		})

		// Should still not throw an error
		assert.NoError(t, jsonData.Clone(t.TempDir()))

		// Make sure the URL was left unresolved
		invalidMatched := false
		_ = jsonData.Walk(func(path string, reader io.Reader) error {
			// Should not have resolved the URL
			if path == "invalid" {
				invalidMatched = true
				data, err := io.ReadAll(reader)
				assert.NoError(t, err)
				assert.Equal(t, string(data), "https://raw.githubusercontent.com/leaktk/fake-leaks/main/this-url-doesnt-exist-8UaehX5b24MzZiaeJ428FK5R")
			}

			return nil
		})

		// Confirm the invalid path was still found
		assert.True(t, invalidMatched, "the invalid path was never checked")
	})
}
