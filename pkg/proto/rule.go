package proto

// Rule that triggered the result
type Rule struct {
	ID          string   `json:"id" toml:"id" yaml:"id"`
	Description string   `json:"description" toml:"description" yaml:"description"`
	Tags        []string `json:"tags" toml:"tags" yaml:"tags"`
}
