package proto

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/leaktk/leaktk/pkg/logger"
)

// RequestKind provides an enum for setting Kind on request
type RequestKind int

const (
	ContainerImageRequestKind RequestKind = iota
	FilesRequestKind
	GitRepoRequestKind
	JSONDataRequestKind
	TextRequestKind
	URLRequestKind
)

// Request is a request to LeakTK
type Request struct {
	ID       string
	Kind     RequestKind
	Resource string
	Opts     Opts
}

// UnmarshalJSON sets r to a copy of data
func (r *Request) UnmarshalJSON(data []byte) error {
	if r == nil {
		return errors.New("Request: UnmarshalJSON on nil pointer")
	}

	var tmp struct {
		ID       string          `json:"id"`
		Kind     string          `json:"kind"`
		Resource string          `json:"resource"`
		Opts     json.RawMessage `json:"options"`
	}

	if err := json.Unmarshal(data, &tmp); err != nil {
		logger.Debug("Request:\n%v", data)
		return fmt.Errorf("could not unmarshal request: %w", err)
	}

	r.ID = tmp.ID
	r.Resource = tmp.Resource

	switch tmp.Kind {
	case "GitRepo":
		r.Kind = GitRepoRequestKind
		r.Opts = &GitRepoOpts{}
	case "JSONData":
		r.Kind = JSONDataRequestKind
		r.Opts = &JSONDataOpts{}
	case "Files":
		r.Kind = FilesRequestKind
		r.Opts = &CommonOpts{}
	case "Text":
		r.Kind = TextRequestKind
		r.Opts = &CommonOpts{}
	case "URL":
		r.Kind = URLRequestKind
		r.Opts = &URLOpts{}
	case "ContainerImage":
		r.Kind = ContainerImageRequestKind
		r.Opts = &ContainerImageOpts{}
	default:
		return fmt.Errorf("unsupported request kind: kind=%q", tmp.Kind)
	}

	if len(tmp.Opts) > 0 {
		if err := json.Unmarshal(tmp.Opts, r.Opts); err != nil {
			logger.Debug("%s:\n%v", tmp.Kind, tmp.Opts)
			return fmt.Errorf("could not unmarshal %s: %w", tmp.Kind, err)
		}
	}

	return nil
}

// Error for returning in the response instead of results if there was a
// critical error causing the scan to fail
type Error struct {
	Code    int    `json:"code"           toml:"code"           yaml:"code"`
	Message string `json:"message"        toml:"message"        yaml:"message"`
	Data    any    `json:"data,omitempty" toml:"data,omitempty" yaml:"data,omitempty"`
}

// Response from the scanner with the scan result
type Response struct {
	ID     string    `json:"id"               toml:"id"               yaml:"id"`
	Result []*Result `json:"result,omitempty" toml:"result,omitempty" yaml:"result,omitempty"`
	Error  *Error    `json:"error,omitempty"  toml:"error,omitempty"  yaml:"data,omitempty"`
}

// Opts is the interface all request options must impleent
type Opts interface {
	Priority() int
}

// CommonOpts provides the baseline options for all resources
type CommonOpts struct {
	priority int `json:"priority"`
}

// Priority tells the priority queue to move this request ahead of things with a
// lower priority
func (o *CommonOpts) Priority() int {
	return o.priority
}

// ContainerImageOpts provides options for a resource of the same name
type ContainerImageOpts struct {
	CommonOpts
	Arch       string   `json:"arch"`
	Depth      int      `json:"depth"`
	Exclusions []string `json:"exclusions"`
	Since      string   `json:"since"`
}

// GitRepoOpts provides options for a resource of the same name
type GitRepoOpts struct {
	CommonOpts
	Branch   string `json:"branch"`
	Depth    int    `json:"depth"`
	Local    bool   `json:"local"`
	Staged   bool   `json:"staged"`
	Since    string `json:"since"`
	Proxy    string `json:"proxy"`
	Unstaged bool   `json:"unstaged"`
}

// JSONDataOpts provides options for a resource of the same name
type JSONDataOpts struct {
	CommonOpts
	FetchURLs string `json:"fetch_urls"`
}

// URLOpts provides options for a resource of the same name
type URLOpts struct {
	CommonOpts
	FetchURLs string `json:"fetch_urls"`
}

// In the future we might have things like GitCommitMessage
// GithubPullRequest, etc
const (
	GenericResultKind          = "Generic"
	ContainerLayerResultKind   = "ContainerLayer"
	ContainerMetdataResultKind = "ContainerMetdata"
	GitCommitResultKind        = "GitCommit"
)

// Result of a scan
type Result struct {
	ID       string            `json:"id" toml:"id" yaml:"id"`
	Kind     string            `json:"kind" toml:"kind" yaml:"kind"`
	Secret   string            `json:"secret" toml:"secret" yaml:"secret"`
	Match    string            `json:"match" toml:"match" yaml:"match"`
	Context  string            `json:"context" toml:"context" yaml:"context"`
	Entropy  float32           `json:"entropy" toml:"entropy" yaml:"entropy"`
	Date     string            `json:"date" toml:"date" yaml:"date"`
	Rule     Rule              `json:"rule" toml:"rule" yaml:"rule"`
	Contact  Contact           `json:"contact" toml:"contact" yaml:"contact"`
	Location Location          `json:"location" toml:"location" yaml:"location"`
	Notes    map[string]string `json:"notes" toml:"notes" yaml:"notes"`
}

// Rule that triggered the result
type Rule struct {
	ID          string   `json:"id" toml:"id" yaml:"id"`
	Description string   `json:"description" toml:"description" yaml:"description"`
	Tags        []string `json:"tags" toml:"tags" yaml:"tags"`
}

// Contact for some resource when available
type Contact struct {
	Name  string `json:"name"  toml:"name"  yaml:"name"`
	Email string `json:"email" toml:"email" yaml:"email"`
}

// Location in the specific resource being scanned
type Location struct {
	Version string `json:"version" toml:"version" yaml:"version"`
	Path    string `json:"path"    toml:"path"    yaml:"path"`
	Start   Point  `json:"start"   toml:"start"   yaml:"start"`
	End     Point  `json:"end"     toml:"end"     yaml:"end"`
}

// Point just provides line & column coordinates for a Result in a text file
type Point struct {
	Line   int `json:"line"   toml:"line"   yaml:"line"`
	Column int `json:"column" toml:"column" yaml:"column"`
}
