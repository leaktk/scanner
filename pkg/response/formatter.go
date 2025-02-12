package response

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v3"

	"github.com/leaktk/scanner/pkg/config"
	"github.com/leaktk/scanner/pkg/logger"
)

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

type Formatter struct {
	format OutputFormat
}

// NewFormatter creates new formatter
func NewFormatter(cfg config.Formatter) (*Formatter, error) {
	format, err := GetOutputFormat(cfg.Format)
	if err != nil {
		return nil, err
	}
	return &Formatter{format: format}, nil
}

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
func (f *Formatter) SetFormat(format OutputFormat) {
	f.format = format
}

// Format renders a response structure to the set format as a string
func (f *Formatter) Format(r *Response) string {
	var output string
	switch f.format {
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
