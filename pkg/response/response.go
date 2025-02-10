package response

import (
	"bytes"
	"encoding/json"
	"fmt"

	"strings"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v3"

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
		ID        string         `json:"id" toml:"id" yaml:"id"`
		Logs      []logger.Entry `json:"logs" toml:"logs" yaml:"logs"`
		RequestID string         `json:"request_id" toml:"request_id" yaml:"request_id"`
		Results   []*Result      `json:"results" toml:"results" yaml:"results"`
	}

	// Result of a scan
	Result struct {
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

	// Location in the specific resource being scanned
	Location struct {
		// This can be things like a commit or some other version control identifier
		Version string `json:"version" toml:"version" yaml:"version"`
		Path    string `json:"path" toml:"path" yaml:"path"`
		// If the start column isn't available it will be zero.
		Start Point `json:"start" toml:"start" yaml:"start"`
		// If the end information isn't available it will be the same as the
		// start information but the column will be the end of the line
		End Point `json:"end" toml:"end" yaml:"end"`
	}

	// Point just provides line & column coordinates for a Result in a text file
	Point struct {
		Line   int `json:"line" toml:"line" yaml:"line"`
		Column int `json:"column" toml:"column" yaml:"column"`
	}

	// Rule that triggered the result
	Rule struct {
		ID          string   `json:"id" toml:"id" yaml:"id"`
		Description string   `json:"description" toml:"description" yaml:"description"`
		Tags        []string `json:"tags" toml:"tags" yaml:"tags"`
	}

	// Contact for some resource when available
	Contact struct {
		Name  string `json:"name" toml:"name" yaml:"name"`
		Email string `json:"email" toml:"email" yaml:"email"`
	}
)

// Parts under here are for formatting the output of a response

type OutputFormat int

const (
	// JSON displays the output in JSON format
	JSON OutputFormat = iota
	// HUMAN displays the outut in a way that's nice for humans to read
	HUMAN
	// TOML displays the output in TOML format
	TOML
	// YAML displays the output in YAML format
	YAML
)

var currentFormat = JSON

// GetOutputFormat takes the string and returns OutputFormat or error
func GetOutputFormat(format string) (OutputFormat, error) {
	format = strings.ToUpper(format)
	switch format {
	case "JSON":
		return JSON, nil
	case "HUMAN":
		return HUMAN, nil
	case "TOML":
		return TOML, nil
	case "YAML":
		return YAML, nil
	default:
		return JSON, fmt.Errorf("invalid output format option: format=%q", format)
	}
}

// SetFormat takes the string for required format
func (*Response) SetFormat(format OutputFormat) {
	currentFormat = format
}

// String renders a response structure to the JSON format
func (r *Response) String() string {
	var output string
	switch currentFormat {
	case JSON:
		output = outputJson(r)
	case HUMAN:
		output = outputHuman(r)
	case TOML:
		output = outputToml(r)
	case YAML:
		output = outputYaml(r)
	}
	return output
}

func outputJson(r *Response) string {
	out, err := json.Marshal(r)
	if err != nil {
		logger.Error("could not marshal response: error=%q", err)
	}
	return string(out)
}

func outputHuman(r *Response) string {
	fmt.Println("Clean it up!")
	return ""
}

func outputToml(r *Response) string {
	var buf bytes.Buffer

	if err := toml.NewEncoder(&buf).Encode(r); err != nil {
		logger.Error("could not marshal response: error=%q", err)
	}
	return buf.String()
}

func outputYaml(r *Response) string {
	out, err := yaml.Marshal(r)
	if err != nil {
		logger.Error("could not marshal response: error=%q", err)
	}
	return string(out)
}
