package resource

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/leaktk/scanner/pkg/response"

	"github.com/leaktk/scanner/pkg/id"
	"github.com/leaktk/scanner/pkg/logger"
)

// WalkFunc is the func signature for functions passed into the various
// resource Walk methods.
type WalkFunc func(path string, reader io.Reader) error

// Resource provides a standard interface for acting with resources in the
// scanner
type Resource interface {
	Clone(path string) error
	ClonePath() string
	Critical(code logger.ErrorCode, msg string, args ...any)
	Debug(code logger.ErrorCode, msg string, args ...any)
	Depth() uint16
	EnrichResult(result *response.Result) *response.Result
	Error(code logger.ErrorCode, msg string, args ...any)
	ID() string
	Info(code logger.ErrorCode, msg string, args ...any)
	Kind() string
	Logs() []logger.Entry
	Priority() int
	ReadFile(path string) ([]byte, error)
	SetCloneTimeout(timeout time.Duration)
	SetDepth(depth uint16)
	SetResourceLogs(enabled bool)
	Since() string
	String() string
	// Walk is the main way to pick through resource data (except for GitRepo)
	Walk(WalkFunc) error
	Warning(code logger.ErrorCode, msg string, args ...any)
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

		if len(options) > 0 {
			if err := json.Unmarshal(options, &gitRepoOptions); err != nil {
				logger.Debug("GitOptions:\n%v", options)
				return nil, fmt.Errorf("could not unmarshal GitOptions: error=%q", err)
			}
		}

		return NewGitRepo(resource, &gitRepoOptions), nil

	case "JSONData":
		var jsonDataOptions JSONDataOptions

		if len(options) > 0 {
			if err := json.Unmarshal(options, &jsonDataOptions); err != nil {
				logger.Debug("JSONDataOptions:\n%v", options)
				return nil, fmt.Errorf("could not unmarshal JSONDataOptions: error=%q", err)
			}
		}

		return NewJSONData(resource, &jsonDataOptions), nil
	case "Files":
		var filesOptions FilesOptions

		if len(options) > 0 {
			if err := json.Unmarshal(options, &filesOptions); err != nil {
				logger.Debug("FilesOptions:\n%v", options)
				return nil, fmt.Errorf("could not unmarshal FilesOptions: error=%q", err)
			}
		}

		return NewFiles(resource, &filesOptions), nil
	case "URL":
		var urlOptions URLOptions

		if len(options) > 0 {
			if err := json.Unmarshal(options, &urlOptions); err != nil {
				logger.Debug("URLOptions:\n%v", options)
				return nil, fmt.Errorf("could not unmarshal URLOptions: error=%q", err)
			}
		}

		return NewURL(resource, &urlOptions), nil
	case "ContainerImage":
		var containerOptions ContainerImageOptions

		if len(options) > 0 {
			if err := json.Unmarshal(options, &containerOptions); err != nil {
				logger.Debug("ContainerImageOptions:\n%v", options)
				return nil, fmt.Errorf("could not unmarshal ContainerImageOptions: error=%q", err)
			}
		}

		return NewContainerImage(resource, &containerOptions), nil
	default:
		return nil, fmt.Errorf("unsupported kind: kind=%q", kind)
	}
}

// BaseResource is a mixin to handle some of the common resource related methods
type BaseResource struct {
	id             string
	logs           []logger.Entry
	LogsInResponse bool
}

// ID returns a path-safe, unique id for this resource
func (r *BaseResource) ID() string {
	if len(r.id) == 0 {
		r.id = id.ID()
	}

	return r.id
}

func (r *BaseResource) Logs() []logger.Entry {
	return r.logs
}

// Critical forwards to the logger and adds to the resource logs used for critical errors that interupt
// the scanner flow.
func (r *BaseResource) Critical(code logger.ErrorCode, msg string, args ...any) {
	if entry := logger.Error(msg, args...); entry != nil {
		entry.Code = code.String()
		entry.Severity = "CRITICAL"
		r.logs = append(r.logs, *entry)
	}
}

// Debug forwards to the logger and adds to the resource logs based on log level
func (r *BaseResource) Debug(code logger.ErrorCode, msg string, args ...any) {
	if entry := logger.Debug(msg, args...); entry != nil {
		entry.Code = code.String()
		if r.LogsInResponse {
			r.logs = append(r.logs, *entry)
		}
	}
}

// Error forwards to the logger and adds to the resource logs based on log level
func (r *BaseResource) Error(code logger.ErrorCode, msg string, args ...any) {
	if entry := logger.Error(msg, args...); entry != nil {
		entry.Code = code.String()
		if r.LogsInResponse {
			r.logs = append(r.logs, *entry)
		}
	}
}

// Warning forwards to the logger and adds to the resource logs based on log level
func (r *BaseResource) Warning(code logger.ErrorCode, msg string, args ...any) {
	if entry := logger.Warning(msg, args...); entry != nil {
		entry.Code = code.String()
		if r.LogsInResponse {
			r.logs = append(r.logs, *entry)
		}
	}
}

// Info forwards to the logger and adds to the resource logs based on log level
func (r *BaseResource) Info(code logger.ErrorCode, msg string, args ...any) {
	if entry := logger.Info(msg, args...); entry != nil {
		entry.Code = code.String()
		if r.LogsInResponse {
			r.logs = append(r.logs, *entry)
		}
	}
}

// SetResourceLogs sets whether to include non-CRITICAL logs
func (r *BaseResource) SetResourceLogs(enabled bool) {
	r.LogsInResponse = enabled
}
