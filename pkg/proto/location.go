package proto

// Location in the specific resource being scanned
type Location struct {
	Version string `json:"version" toml:"version" yaml:"version"`
	Path    string `json:"path"    toml:"path"    yaml:"path"`
	Start   Point  `json:"start"   toml:"start"   yaml:"start"`
	End     Point  `json:"end"     toml:"end"     yaml:"end"`
}
