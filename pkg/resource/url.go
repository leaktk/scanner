package resource

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	httpclient "github.com/leaktk/leaktk/pkg/http"
	"github.com/leaktk/leaktk/pkg/id"
)

// TODO: just scan this with a gitleaks reader or JSON source on the fly
func CloneURL(url, path string) error {
	client := httpclient.NewClient()
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("http GET error: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: status_code=%d", resp.StatusCode)
	}
	defer resp.Body.Close()

	if strings.HasPrefix(resp.Header.Get("Content-Type"), "application/json") {
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("could not read JSON response body: %w", err)
		}

		// Scan as a JSONData resource
		r.resource = NewJSONData(string(data), &JSONDataOptions{})
	} else {
		if err := os.MkdirAll(r.path, 0700); err != nil {
			return fmt.Errorf("could not create path: path=%q error=%q", r.path, err)
		}

		dataPath := filepath.Clean(filepath.Join(r.path, id.ID(r.url)))
		dataFile, err := os.OpenFile(dataPath, os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			return fmt.Errorf("could not open data file: error=%q", err)
		}
		defer dataFile.Close()

		// Store the file on disk for scanning
		_, err = io.Copy(dataFile, resp.Body)
		if err != nil {
			return fmt.Errorf("could not copy data: error=%q", err)
		}
		// Scan as a Files resource
		r.resource = NewFiles(dataPath, &FilesOptions{})
	}

	return r.resource.Clone(r.path)
}
