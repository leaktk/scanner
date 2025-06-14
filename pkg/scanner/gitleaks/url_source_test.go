package gitleaks

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zricethezav/gitleaks/v8/sources"
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

	// Test general content
	generalURL, err := url.JoinPath(ts.URL, "general")
	assert.NoError(t, err)

	source := URL{
		RawURL: generalURL,
	}

	fragments := []sources.Fragment{}
	source.Fragments(context.Background(), func(fragment sources.Fragment, err error) error {
		fragments = append(fragments, fragment)
		return nil
	})

	assert.Len(t, fragments, 1)
	assert.Equal(t, fragments[0].FilePath, "/general")
	assert.Equal(t, fragments[0].Raw, "general-content")

	// Test json data
	jsonDataURL, err := url.JoinPath(ts.URL, "data.json")
	assert.NoError(t, err)
	source = URL{
		RawURL: jsonDataURL,
	}

	fragments = []sources.Fragment{}
	source.Fragments(context.Background(), func(fragment sources.Fragment, err error) error {
		fragments = append(fragments, fragment)
		return nil
	})

	assert.Len(t, fragments, 1)
	assert.Equal(t, fragments[0].FilePath, "/data.json!data")
	assert.Equal(t, fragments[0].Raw, "json-data")
}
