package proto

// Contact for some resource when available
type Contact struct {
	Name  string `json:"name"  toml:"name"  yaml:"name"`
	Email string `json:"email" toml:"email" yaml:"email"`
}
