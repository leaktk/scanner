package resource

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	httpclient "github.com/leaktk/scanner/pkg/http"
	"github.com/leaktk/scanner/pkg/id"
	"github.com/leaktk/scanner/pkg/response"
)

// URL provides a to pull remote content by a URL
type URL struct {
	// Provide common helper functions
	BaseResource
	clonePath string
	resource  Resource
	url       string
	options   *URLOptions
}

// URLOptions are options for the URL resource
type URLOptions struct {
	// The scan priority
	Priority int `json:"priority"`
}

// NewURL returns a configured URL resource for the scanner to scan
func NewURL(url string, options *URLOptions) *URL {
	return &URL{
		url:     url,
		options: options,
	}
}

// Kind of resource (always returns URL here)
func (r *URL) Kind() string {
	return "URL"
}

// String representation of the resource
func (r *URL) String() string {
	return r.url
}

// Clone the resource to the desired local location and store the path
func (r *URL) Clone(path string) error {
	r.clonePath = path

	client := httpclient.NewClient()
	resp, err := client.Get(r.url)
	if err != nil {
		return fmt.Errorf("http GET error: error=%q", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: status_code=%d", resp.StatusCode)
	}
	defer resp.Body.Close()

	if strings.HasPrefix(resp.Header.Get("Content-Type"), "application/json") {
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("could not read JSON response body: error=%q", err)
		}

		// Scan as a JSONData resource
		r.resource = NewJSONData(string(data), &JSONDataOptions{})
	} else {
		if err := os.MkdirAll(r.clonePath, 0700); err != nil {
			return fmt.Errorf("could not create clone path: path=%q error=%q", r.clonePath, err)
		}

		// Use ID(r.url) to create the dataPath so that we don't have collisions
		// for if we follow URLs in the returned content
		dataPath := filepath.Clean(filepath.Join(r.clonePath, id.ID(r.url)))
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

	return r.resource.Clone(r.clonePath)
}

// Path returns where this repo has been cloned if cloned else ""
func (r *URL) Path() string {
	return r.clonePath
}

// Depth returns the depth for things that have version control
func (r *URL) Depth() uint16 {
	return 0
}

// EnrichResult enriches the result with contextual information
func (r *URL) EnrichResult(result *response.Result) *response.Result {
	result.Kind = response.GeneralResultKind
	return result
}

// SetDepth allows you to adjust the depth for the resource
func (r *URL) SetDepth(depth uint16) {
	// no-op
}

// SetCloneTimeout lets you adjust the timeout before the clone aborts
func (r *URL) SetCloneTimeout(timeout time.Duration) {
	// no-op
}

// Since returns the date after which things should be scanned for things
// that have versions
func (r *URL) Since() string {
	return ""
}

// ReadFile provides a way to access values in the resource
func (r *URL) ReadFile(path string) ([]byte, error) {
	return r.resource.ReadFile(path)
}

// Walk traverses the resource like a directory tree
func (r *URL) Walk(fn WalkFunc) error {
	return r.resource.Walk(fn)
}

// Priority returns the scan priority
func (r *URL) Priority() int {
	return r.options.Priority
}

// IsLocal returns whether this is a local resource or not
func (r *URL) IsLocal() bool {
	return false
}
