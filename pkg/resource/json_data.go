package resource

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/leaktk/scanner/pkg/logger"
)

// JSONData provides a way to interact with a json data as a resource
type JSONData struct {
	// Provide common helper functions
	BaseResource
	raw       string
	data      any
	clonePath string
	options   *JSONDataOptions
}

// JSONDataOptions are options for the JSONData resource
type JSONDataOptions struct {
	// Currently none needed but here for future cases
}

// NewJSONData returns a configured git repo resource for the scanner to scan
func NewJSONData(raw string, options *JSONDataOptions) *JSONData {
	return &JSONData{
		raw:     raw,
		options: options,
	}
}

// Kind of resource (always returns JSONData here)
func (r *JSONData) Kind() string {
	return "JSONData"
}

// String representation of the resource
func (r *JSONData) String() string {
	return r.raw
}

// Clone the resource to the desired local location and store the path
func (r *JSONData) Clone(path string) error {
	r.clonePath = path
	if err := os.MkdirAll(r.clonePath, 0700); err != nil {
		return err
	}

	// Load the raw json into the data variable
	if err := json.Unmarshal([]byte(r.raw), &r.data); err != nil {
		logger.Debug("JSONData:\n%v", r.raw)
		return fmt.Errorf("could not unmarshal JSONData: error=%q", err)
	}

	// Support droping specific files in the "repo" to configure scanner backends
	for _, file := range []string{
		".gitleaks.toml",
		".gitleaksignore",
		".gitleaksbaseline",
	} {
		if data, err := r.ReadFile(file); err == nil {
			if err := os.WriteFile(filepath.Join(r.clonePath, file), data, 0600); err != nil {
				return err
			}
		}
	}

	return nil
}

// ClonePath returns where this repo has been cloned if cloned else ""
func (r *JSONData) ClonePath() string {
	return r.clonePath
}

// Depth returns the depth for things that have version control
func (r *JSONData) Depth() uint16 {
	return 0
}

// SetDepth allows you to adjust the depth for the resource
func (r *JSONData) SetDepth(depth uint16) {
	// no-op
}

// SetCloneTimeout lets you adjust the timeout before the clone aborts
func (r *JSONData) SetCloneTimeout(timeout time.Duration) {
	// no-op
}

// Since returns the date after which things should be scanned for things
// that have versions
func (r *JSONData) Since() string {
	return ""
}

// ReadFile provides a way to access values in the JSON data
func (r *JSONData) ReadFile(path string) ([]byte, error) {
	// Build out the path components
	var components []string
	for _, component := range strings.Split(path, string(filepath.Separator)) {
		if component == "" {
			continue
		}
		components = append(components, component)
	}

	// Traverse the data structure
	doesNotExistError := fmt.Errorf("%q does not exist", path)
	current := r.data
	for _, component := range components {
		switch obj := current.(type) {
		case []any:
			i, err := strconv.Atoi(component)

			if err != nil {
				return []byte{}, fmt.Errorf("%q must be an integer", component)
			}

			if len(obj) <= i {
				return []byte{}, doesNotExistError
			}

			current = obj[i]
		case map[string]any:
			var ok bool

			current, ok = obj[component]

			if !ok {
				return []byte{}, doesNotExistError
			}
		default:
			return []byte{}, doesNotExistError
		}
	}

	// Look at the value of current after traversal and return it if reached
	switch current.(type) {
	case map[string]any:
		return []byte{}, doesNotExistError
	case []any:
		return []byte{}, doesNotExistError
	case nil: // Handle nil
		return []byte{}, nil
	default: // Handle bool, float64, and string
		return []byte(fmt.Sprintf("%v", current)), nil
	}
}

func (r *JSONData) walkRecusrive(path string, current any, fn WalkFunc) error {
	switch obj := current.(type) {
	case map[string]any:
		for key, value := range obj {
			subPath := filepath.Join(path, key)

			if err := r.walkRecusrive(subPath, value, fn); err != nil {
				return err
			}
		}
		return nil
	case []any:
		for i, value := range obj {
			subPath := filepath.Join(path, strconv.Itoa(i))

			if err := r.walkRecusrive(subPath, value, fn); err != nil {
				return err
			}
		}
		return nil
	case nil: // Handle nil
		return fn(path, []byte{})
	default: // Handle bool, float64, and string
		return fn(path, []byte(fmt.Sprintf("%v", obj)))
	}
}

// Walk traverses the JSON data structure like it's a directory tree
func (r *JSONData) Walk(fn WalkFunc) error {
	return r.walkRecusrive("", r.data, fn)
}
