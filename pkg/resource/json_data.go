package resource

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/leaktk/scanner/pkg/fs"
	"github.com/leaktk/scanner/pkg/logger"
	"github.com/leaktk/scanner/pkg/response"
)

var urlRegexp = regexp.MustCompile(`^https?:\/\/\S+$`)

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
	// *:/foo/*/bar*:/foo
	FetchURLs string `json:"fetch_urls"`
	// The scan priority
	Priority int `json:"priority"`
}

type jsonNode struct {
	parent any
	key    string
	path   string
	value  any
}

type jsonWalkFunc func(leafNode jsonNode) error

// NewJSONData returns a configured JSONData resource for the scanner to scan
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
	var err error

	r.clonePath = path

	if err = os.MkdirAll(r.clonePath, 0700); err != nil {
		return err
	}

	// Load the raw json into the data variable
	if err = json.Unmarshal([]byte(r.raw), &r.data); err != nil {
		r.Debug(logger.ScanDetail, "JSONData:\n%v", r.raw)
		return fmt.Errorf("could not unmarshal JSONData: %w", err)
	}

	// Support dropping specific files in the "repo" to configure scanner backends
	for _, file := range []string{
		".gitleaks.toml",
		".gitleaksignore",
		".gitleaksbaseline",
	} {
		if data, err := r.ReadFile(file); err == nil {
			if err = os.WriteFile(filepath.Join(r.clonePath, file), data, 0600); err != nil {
				return err
			}
		}
	}

	// Fetch URLs in jsonNodes and replace the node with a resource object
	if len(r.options.FetchURLs) > 0 {
		err = r.fetchURLs(jsonNode{value: r.data}, r.clonePath)
	}

	return err
}

func (r *JSONData) shouldFetchURL(path string) bool {
	for _, pattern := range strings.Split(r.options.FetchURLs, ":") {
		if fs.Match(pattern, path) {
			return true
		}
	}

	return false
}

func (r *JSONData) fetchURLs(rootNode jsonNode, clonePath string) error {
	return r.walkRecusrive(rootNode, func(leafNode jsonNode) error {
		// We only want string objects
		obj, isString := leafNode.value.(string)
		if !isString {
			return nil
		}

		if !urlRegexp.MatchString(obj) {
			return nil
		}

		if !r.shouldFetchURL(leafNode.path) {
			r.Debug(logger.CloneDetail, "not fetching URL path=%q url=%q", leafNode.path, obj)
			return nil
		}

		urlResource := NewURL(obj, &URLOptions{})
		r.Info(logger.CloneDetail, "fetching url url=%q", obj)
		err := urlResource.Clone(filepath.Join(clonePath, leafNode.path))

		if err != nil {
			// Not being able to retrieve a URL found inside JSONData is not a fatal error. Logging until update how
			// we manage fatal/nonfatal errors flowing through the application.
			r.Error(logger.CloneError, "could not fetch url: %w path=%q url=%q", err, leafNode.path, obj)
			return nil
		}

		r.Info(logger.CloneDetail, "replacing url with contents path=%q url=%q", leafNode.path, obj)
		return r.replaceWithResource(leafNode, urlResource)
	})
}

// ClonePath returns where this repo has been cloned if cloned else ""
func (r *JSONData) ClonePath() string {
	return r.clonePath
}

// Depth returns the depth for things that have version control
func (r *JSONData) Depth() uint16 {
	return 0
}

// EnrichResult enriches the result with contextual information
func (r *JSONData) EnrichResult(result *response.Result) *response.Result {
	result.Kind = response.JSONDataResultKind
	return result
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
	doesNotExistError := fmt.Errorf("path does not exist path=%q", path)
	current := r.data
	for i, component := range components {
		switch obj := current.(type) {
		case []any:
			i, err := strconv.Atoi(component)

			if err != nil {
				return []byte{}, fmt.Errorf("component must be an integer component=%q", component)
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
		case Resource:
			return obj.ReadFile(filepath.Join(components[i:]...))
		default:
			return []byte{}, doesNotExistError
		}
	}

	// Look at the value of current after traversal and return it if reached
	switch obj := current.(type) {
	case map[string]any:
		return []byte{}, doesNotExistError
	case []any:
		return []byte{}, doesNotExistError
	case nil: // Handle nil
		return []byte{}, nil
	case Resource:
		return obj.ReadFile("")
	default: // Handle bool, float64, and string
		return []byte(fmt.Sprintf("%v", obj)), nil
	}
}

func (r *JSONData) walkRecusrive(currentNode jsonNode, fn jsonWalkFunc) error {
	switch obj := currentNode.value.(type) {
	case map[string]any:
		for key, value := range obj {
			childNode := jsonNode{
				parent: currentNode.value,
				key:    key,
				path:   filepath.Join(currentNode.path, key),
				value:  value,
			}

			if err := r.walkRecusrive(childNode, fn); err != nil {
				return err
			}
		}

		return nil
	case []any:
		for i, value := range obj {
			key := strconv.Itoa(i)

			childNode := jsonNode{
				parent: currentNode.value,
				key:    key,
				path:   filepath.Join(currentNode.path, key),
				value:  value,
			}

			if err := r.walkRecusrive(childNode, fn); err != nil {
				return err
			}
		}
		return nil
	default: // We found a leaf node
		return fn(currentNode)
	}
}

// Take a leaf node in the JSON tree and replace it with a resource object
func (r *JSONData) replaceWithResource(leafNode jsonNode, resource Resource) error {
	switch parent := leafNode.parent.(type) {
	case map[string]any:
		parent[leafNode.key] = resource
	case []any:
		i, err := strconv.Atoi(leafNode.key)

		if err != nil {
			return fmt.Errorf("could not set resource: %w path=%q", err, leafNode.path)
		}

		parent[i] = resource
	default:
		// Not sure how you got here
		return fmt.Errorf(`leaf node parent was a leaf node some how: ¯\_(ツ)_/¯`)
	}

	return nil
}

// prefixClonePath handles providing the full clone path for sub-resources.
// When a node in the tree is replaced with a resource, the resource isn't
// aware of its place in the tree when you call Walk on it. This adds that path
// back.
func (r *JSONData) prefixClonePath(leafNode jsonNode, fn WalkFunc) WalkFunc {
	return func(path string, reader io.Reader) error {
		return fn(filepath.Join(leafNode.path, path), reader)
	}
}

// walkFuncToJSONWalkFunc takes a normal WalkFunc and wraps it in a
// jsonWalkFunc so it can be used in this resource. The custom jsonWalkFunc
// exists since there are multiple cases where we need to walk through the json
// data structure that wouldn't apply to other resources.
func (r *JSONData) walkFuncToJSONWalkFunc(fn WalkFunc) jsonWalkFunc {
	return func(leafNode jsonNode) error {
		switch obj := leafNode.value.(type) {
		case nil: // Handle nil
			return fn(leafNode.path, bytes.NewReader([]byte{}))
		case Resource:
			return obj.Walk(r.prefixClonePath(leafNode, fn))
		default: // Handle bool, float64, and string
			return fn(leafNode.path, bytes.NewReader([]byte(fmt.Sprintf("%v", obj))))
		}
	}
}

// Walk traverses the JSON data structure like it's a directory tree
func (r *JSONData) Walk(fn WalkFunc) error {
	return r.walkRecusrive(jsonNode{value: r.data}, r.walkFuncToJSONWalkFunc(fn))
}

// Priority returns the scan priority
func (r *JSONData) Priority() int {
	return r.options.Priority
}
