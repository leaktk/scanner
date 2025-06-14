package proto

// Response from the scanner with the scan results
type Response struct {
	ID        string    `json:"id"         toml:"id"         yaml:"id"`
	RequestID string    `json:"request_id" toml:"request_id" yaml:"request_id"`
	Results   []*Result `json:"results"    toml:"results"    yaml:"results"`
}
