package response

import (
	"encoding/json"

	"github.com/leaktk/scanner/pkg/logger"
)

// In the future we might have things like GitCommitMessage
// GithubPullRequest, etc
const (
	GeneralResultKind          = "General"
	GitCommitResultKind        = "GitCommit"
	JSONDataResultKind         = "JSONData"
	ContainerLayerResultKind   = "ContainerLayer"
	ContainerMetdataResultKind = "ContainerMetdata"
)

type (
	// Response from the scanner with the scan results
	Response struct {
		ID        string        `json:"id"`
		Errors    []LeakTKError `json:"errors"`
		RequestID string        `json:"request_id"`
		Results   []*Result     `json:"results"`
	}

	// Result of a scan
	Result struct {
		ID       string            `json:"id"`
		Kind     string            `json:"kind"`
		Secret   string            `json:"secret"`
		Match    string            `json:"match"`
		Context  string            `json:"context"`
		Entropy  float32           `json:"entropy"`
		Date     string            `json:"date"`
		Rule     Rule              `json:"rule"`
		Contact  Contact           `json:"contact"`
		Location Location          `json:"location"`
		Notes    map[string]string `json:"notes"`
	}

	// Location in the specific resource being scanned
	Location struct {
		// This can be things like a commit or some other version control identifier
		Version string `json:"version"`
		Path    string `json:"path"`
		// If the start column isn't available it will be zero.
		Start Point `json:"start"`
		// If the end information isn't available it will be the same as the
		// start information but the column will be the end of the line
		End Point `json:"end"`
	}

	// Point just provides line & column coordinates for a Result in a text file
	Point struct {
		Line   int `json:"line"`
		Column int `json:"column"`
	}

	// Rule that triggered the result
	Rule struct {
		ID          string   `json:"id"`
		Description string   `json:"description"`
		Tags        []string `json:"tags"`
	}

	// Contact for some resource when available
	Contact struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}
)

// String renders a response structure to the JSON format
func (r *Response) String() string {
	out, err := json.Marshal(r)
	if err != nil {
		logger.Error("could not marshal response: error=%q", err)
	}

	return string(out)
}
