package resource

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/leaktk/scanner/pkg/id"
)

// WalkFunc is the func signature for functions passed into the various
// resource Walk methods.
type WalkFunc func(path string, data []byte) error

// Resource provides a standard interface for acting with resources in the
// scanner
type Resource interface {
	Clone(path string) error
	ClonePath() string
	Depth() uint16
	ID() string
	Kind() string
	ReadFile(path string) ([]byte, error)
	SetCloneTimeout(timeout time.Duration)
	SetDepth(depth uint16)
	Since() string
	String() string
	// Walk is the main way to pick through resource data (except for GitRepo)
	Walk(WalkFunc) error
}

// NewResource handles building out the resource from kind, the resource string
// and the options as a raw json message
func NewResource(kind, resource string, options json.RawMessage) (Resource, error) {
	// When adding different kinds, make sure to provide the corresponding
	// request.<Kind>Options() (*<Kind>) that returns nil if that isn't the
	// type of options on the object.
	switch kind {
	case "GitRepo":
		var gitRepoOptions GitRepoOptions

		if err := json.Unmarshal(options, &gitRepoOptions); err != nil {
			return nil, err
		}

		return NewGitRepo(resource, &gitRepoOptions), nil

	case "JSONData":
		var jsonDataOptions JSONDataOptions

		if err := json.Unmarshal(options, &jsonDataOptions); err != nil {
			return nil, err
		}

		return NewJSONData(resource, &jsonDataOptions), nil
	default:
		return nil, fmt.Errorf("unsupported kind: kind=%q", kind)
	}
}

// BaseResource is a mixin to handle some of the common resource related methods
type BaseResource struct {
	id string
}

// ID returns a path-safe, unique id for this resource
func (r *BaseResource) ID() string {
	if len(r.id) == 0 {
		r.id = id.ID()
	}

	return r.id
}
