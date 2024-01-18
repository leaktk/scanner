package scanner

// Request to the scanner to scan some resource
type Request struct {
	ID       string            `json:"id"`
	Kind     string            `json:"kind"`
	Resource string            `json:"resource"`
	Options  map[string]string `json:"options"`
}

// NewRequest is for creating request objects manually instead of unmarshaling
// them though JSON
func NewRequest(id, kind, resource string, options map[string]string) Request {
	return Request{
		ID:       id,
		Kind:     kind,
		Resource: resource,
		Options:  options,
	}
}
