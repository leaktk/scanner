package resource

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/leaktk/scanner/pkg/response"
)

// TextKind defines what kind value maps to this resource
const TextKind = "Text"

// Text provides a way to interact with plain text
type Text struct {
	// Provide common helper functions
	BaseResource
	data    string
	path    string
	options *TextOptions
}

// TextOptions are options for the Text resource
type TextOptions struct {
	// The scan priority
	Priority int `json:"priority"`
}

// NewText returns a configured Text resource for the scanner to scan
func NewText(data string, options *TextOptions) *Text {
	return &Text{
		data:    data,
		options: options,
	}
}

// Kind of resource (always returns Text here)
func (r *Text) Kind() string {
	return TextKind
}

// String representation of the resource
func (r *Text) String() string {
	return r.data
}

// Clone the resource to the desired local location and store the path
func (r *Text) Clone(path string) error {
	var err error

	r.path = path

	if err = os.MkdirAll(r.path, 0700); err != nil {
		return err
	}

	return nil
}

// Path returns where this text has been cloned if cloned else ""
func (r *Text) Path() string {
	return r.path
}

// Depth returns the depth for things that have version control
func (r *Text) Depth() uint16 {
	return 0
}

// EnrichResult enriches the result with contextual information
func (r *Text) EnrichResult(result *response.Result) *response.Result {
	result.Kind = response.TextResultKind
	return result
}

// IsLocal returns whether this is a local resource or not
func (r *Text) IsLocal() bool {
	return false
}

// SetDepth allows you to adjust the depth for the resource
func (r *Text) SetDepth(depth uint16) {
	// no-op
}

// SetCloneTimeout lets you adjust the timeout before the clone aborts
func (r *Text) SetCloneTimeout(timeout time.Duration) {
	// no-op
}

// Since returns the date after which things should be scanned for things
// that have versions
func (r *Text) Since() string {
	return ""
}

// ReadFile provides a way to access values in the text data
func (r *Text) ReadFile(path string) ([]byte, error) {
	if len(path) == 0 {
		return []byte(r.data), nil
	}

	return []byte{}, fmt.Errorf("%q does not exist", path)
}

// Walk returns the text as a single item in the "tree"
func (r *Text) Walk(fn WalkFunc) error {
	return fn("", strings.NewReader(r.data))
}

// Priority returns the scan priority
func (r *Text) Priority() int {
	return r.options.Priority
}
