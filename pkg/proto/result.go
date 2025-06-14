package proto

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
