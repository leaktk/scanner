package response

import (
	"bytes"
	"encoding/csv"
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

// Below are for formatting the output of a response, here as specific to the Response

// OutputFormat is the code(int) for each format
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
	// CSV displays the
	CSV
)

// Specifies the default format and stores any changes.
var currentFormat = JSON

// GetOutputFormat takes the string and returns OutputFormat or an error
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
	case "CSV":
		return CSV, nil
	default:
		return JSON, fmt.Errorf("invalid output format option: format=%q", format)
	}
}

// SetFormat sets the format using the OutputFormat value
func (*Response) SetFormat(format OutputFormat) {
	currentFormat = format
}

// String renders a response structure to the set format as a string
func (r *Response) String() string {
	var output string
	switch currentFormat {
	case JSON:
		output = formatJson(r)
	case HUMAN:
		output = formatHuman(r)
	case TOML:
		output = formatToml(r)
	case YAML:
		output = formatYaml(r)
	case CSV:
		output = formatCsv(r)
	}
	return output
}

func formatJson(r *Response) string {
	out, err := json.Marshal(r)
	if err != nil {
		logger.Error("could not marshal response: error=%q", err)
	}
	return string(out)
}

func formatHuman(r *Response) string {
	headers := flattenedResponseFields()
	for i, header := range headers {
		// Append a : to each label. Do it once here instead of every loop
		headers[i] = header + ":"
	}
	flat := flattenedResponse(r)
	var out []string
	for _, response := range flat {
		for i, entry := range response {
			// Specifies width of 26 characters for labels
			out = append(out, fmt.Sprintf("%-26s%s", headers[i], entry))
		}
		out = append(out, "\n")
	}
	return strings.Join(out, "\n")
}

func formatToml(r *Response) string {
	var buf bytes.Buffer

	if err := toml.NewEncoder(&buf).Encode(r); err != nil {
		logger.Error("could not marshal response: error=%q", err)
	}
	return buf.String()
}

func formatYaml(r *Response) string {
	out, err := yaml.Marshal(r)
	if err != nil {
		logger.Error("could not marshal response: error=%q", err)
	}
	return string(out)
}

func formatCsv(r *Response) string {
	headers := flattenedResponseFields()

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	err := writer.Write(headers)
	if err != nil {
		logger.Error("could not write response: error=%q", err)
	}

	err = writer.WriteAll(flattenedResponse(r))
	if err != nil {
		logger.Error("could not write response: error=%q", err)
	}

	return buf.String()
}

// flattenedResponseFields provides a list containing the field labels for a flattened response
func flattenedResponseFields() []string {
	return []string{"ID", "REQUEST.ID", "RESULT.ID", "RESULT.KIND", "RESULT.SECRET", "RESULT.MATCH",
		"RESULT.CONTEXT", "RESULT.ENTROPY", "RESULT.DATE", "RESULT.NOTES", "RESULT.RULE.ID", "RESULT.RULE.DESCRIPTION",
		"RESULT.RULE.TAGS", "RESULT.CONTACT", "RESULT.LOCATION.VERSION", "RESULT.LOCATION.PATH",
		"RESULT.LOCATION.RANGE"}
}

// flattenedResponse takes the response and returns a 2d list of responses in flattenedResponseFields order
func flattenedResponse(response *Response) [][]string {
	var flattened [][]string

	for _, result := range response.Results {
		flattened = append(flattened, []string{
			sanitizeField(response.ID),
			sanitizeField(response.RequestID),
			sanitizeField(result.ID),
			sanitizeField(result.Kind),
			sanitizeField(result.Secret),
			sanitizeField(result.Match),
			sanitizeField(result.Context),
			sanitizeField(fmt.Sprintf("%f", result.Entropy)),
			sanitizeField(result.Date),
			sanitizeField(flattenMapToString(result.Notes)),
			sanitizeField(result.Rule.ID),
			sanitizeField(result.Rule.Description),
			sanitizeField(strings.Join(result.Rule.Tags, ", ")),
			sanitizeField(flattenContact(result.Contact)),
			sanitizeField(result.Location.Version),
			sanitizeField(result.Location.Path),
			sanitizeField(fmt.Sprintf("L%dC%d-L%dC%d", result.Location.Start.Line,
				result.Location.Start.Column, result.Location.End.Line, result.Location.End.Column)),
		})
	}

	return flattened
}

// flattenMapToString creates a flat string with pairs "key: value" only single depth
func flattenMapToString(m map[string]string) string {
	var pairs []string
	for k, v := range m {
		pairs = append(pairs, fmt.Sprintf("%s: %s", k, v))
	}
	return strings.Join(pairs, ", ")
}

// flattenContact creates a single string with Contact information as "Name <Email>"
func flattenContact(contact Contact) string {
	return fmt.Sprintf("%s <%s>", contact.Name, contact.Email)
}

// sanitizeField takes a string and makes it safe for CSV by replacing newlines and escaping quotes
// CSV is not a great format for rich/nested data, this makes the data more consistent for the format
func sanitizeField(value string) string {
	if strings.ContainsAny(value, ",\n\"") {
		value = strings.ReplaceAll(value, "\"", "\"\"")
		value = strings.ReplaceAll(value, "\n", " ")
		return fmt.Sprintf("\"%s\"", value)
	}
	return value
}
