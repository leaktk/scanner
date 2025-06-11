package resource

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestURL(t *testing.T) {
	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var content string

		switch r.URL.Path {
		case "/general":
			w.Header().Add("Content-Type", "text/plain")
			content = "general-content"
		case "/data.json":
			w.Header().Add("Content-Type", "application/json")
			content = "{\"data\": \"json-data\"}"
		default:
			t.Errorf("invalid URL path: path=%q", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		_, err := io.WriteString(w, content)
		assert.NoError(t, err)
	}))

	ts.Start()
	defer ts.Close()

	t.Run("Objects", func(t *testing.T) {
		tsURL, err := url.JoinPath(ts.URL, "general")
		assert.NoError(t, err)

		urlResource := NewURL(tsURL, &URLOptions{})
		err = urlResource.Clone(t.TempDir())
		assert.NoError(t, err)

		_ = urlResource.Objects(func(obj Object) error {
			data, err := io.ReadAll(obj.Content)
			assert.NoError(t, err)
			assert.Equal(t, obj.Path, "")
			assert.Equal(t, string(data), "general-content")
			return nil
		})

		tsURL, err = url.JoinPath(ts.URL, "data.json")
		assert.NoError(t, err)
		urlResource = NewURL(tsURL, &URLOptions{})
		err = urlResource.Clone(t.TempDir())
		assert.NoError(t, err)

		_ = urlResource.Objects(func(obj Object) error {
			data, err := io.ReadAll(obj.Content)
			assert.NoError(t, err)
			assert.Equal(t, obj.Path, "data")
			assert.Equal(t, string(data), "json-data")
			return nil
		})
	})
}
