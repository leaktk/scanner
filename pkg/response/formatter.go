package response

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"slices"
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
	// CSV displays the output in CSV format
	CSV
)

type Formatter struct {
	format   OutputFormat
	truncate int
}

// NewFormatter creates new formatter
func NewFormatter(cfg config.Formatter) (*Formatter, error) {
	format, err := GetOutputFormat(cfg.Format)
	if err != nil {
		return nil, err
	}
	return &Formatter{format: format, truncate: cfg.Truncate}, nil
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

// Format renders a response structure to the set format as a string
func (f *Formatter) Format(r *Response) string {
	var output string
	switch f.format {
	case JSON:
		output = f.formatJson(r)
	case HUMAN:
		output = f.formatHuman(r)
	case TOML:
		output = f.formatToml(r)
	case YAML:
		output = f.formatYaml(r)
	case CSV:
		output = f.formatCsv(r)
	}
	return output
}

func (f *Formatter) formatJson(r *Response) string {
	out, err := json.Marshal(r)
	if err != nil {
		logger.Error("could not marshal response: error=%q", err)
	}
	return string(out)
}

func (f *Formatter) formatHuman(r *Response) string {
	var out strings.Builder
	headers := flattenedResponseFields()
	truncated := []int{}
	if f.truncate > 0 {
		truncated = truncatableResponseFields()
	}
	flat := flattenedResponse(r, false)
	for _, response := range flat {
		for i, entry := range response {
			if slices.Contains(truncated, i) && len(entry) > f.truncate {
				_, _ = fmt.Fprintf(&out, "%-26s: %s...\n", headers[i], entry[:f.truncate])
			} else {
				_, _ = fmt.Fprintf(&out, "%-26s: %s\n", headers[i], entry)
			}
		}
		out.WriteRune('\n')
	}
	return out.String()
}

func (f *Formatter) formatToml(r *Response) string {
	var buf bytes.Buffer

	if err := toml.NewEncoder(&buf).Encode(r); err != nil {
		logger.Error("could not marshal response: error=%q", err)
	}
	return buf.String()
}

func (f *Formatter) formatYaml(r *Response) string {
	out, err := yaml.Marshal(r)
	if err != nil {
		logger.Error("could not marshal response: error=%q", err)
	}
	return string(out)
}

func (f *Formatter) formatCsv(r *Response) string {
	headers := flattenedResponseFields()

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	err := writer.Write(headers)
	if err != nil {
		logger.Error("could not write response: error=%q", err)
	}

	err = writer.WriteAll(flattenedResponse(r, true))
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

// flattenedResponseFields provides a list containing the field labels for a flattened response
func truncatableResponseFields() []int {
	truncatableFields := []string{"RESULT.SECRET", "RESULT.MATCH", "RESULT.CONTEXT"}
	var fields []int

	for i, entry := range flattenedResponseFields() {
		if slices.Contains(truncatableFields, entry) {
			fields = append(fields, i)
		}
	}
	return fields
}

// flattenedResponse takes the response and returns a 2d list of responses in flattenedResponseFields order
func flattenedResponse(response *Response, sanitize bool) [][]string {
	var flattened [][]string

	for _, result := range response.Results {
		entry := []string{
			response.ID,
			response.RequestID,
			result.ID,
			result.Kind,
			result.Secret,
			result.Match,
			result.Context,
			fmt.Sprintf("%f", result.Entropy),
			result.Date,
			flattenMapToString(result.Notes),
			result.Rule.ID,
			result.Rule.Description,
			strings.Join(result.Rule.Tags, ", "),
			flattenContact(result.Contact),
			result.Location.Version,
			result.Location.Path,
			fmt.Sprintf("L%dC%d-L%dC%d", result.Location.Start.Line,
				result.Location.Start.Column, result.Location.End.Line, result.Location.End.Column),
		}
		if sanitize {
			entry = sanitizeEntry(entry)
		}
		flattened = append(flattened, entry)
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
func sanitizeEntry(value []string) []string {
	var output []string
	for _, entry := range value {
		if strings.ContainsAny(entry, ",\n\"") {
			entry = strings.ReplaceAll(entry, "\"", "\"\"")
			entry = strings.ReplaceAll(entry, "\n", " ")
			entry = fmt.Sprintf("\"%s\"", entry)
		}
		output = append(output, entry)
	}
	return output
}
