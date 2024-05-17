package scanner

// Request to the scanner to scan some resource
type Request struct {
	// Client provided identifier for associating a response to a request
	ID string `json:"id"`
	// Kind of thing being scanned
	Kind string `json:"kind"`
	// Thing to scan (e.g. URL, snippet of text, etc)
	Resource string `json:"resource"`
	// Flags to pass to the scanner (these depend heavily on the Kind)
	Options map[string]string `json:"options"`
}

// NewRequest is for creating request objects manually instead of unmarshaling
// them though JSON
func NewRequest(id, kind, resource string, options map[string]string) *Request {
	return &Request{
		ID:       id,
		Kind:     kind,
		Resource: resource,
		Options:  options,
	}
}
