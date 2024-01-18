package scanner

// Scan takes a recquest scans the resource and returns the results response
func Scan(request *Request) (*Response, error) {
	resp := Response{
		Request: RequestDetails{
			ID:       request.ID,
			Kind:     request.Kind,
			Resource: request.Resource,
		},
	}

	// TODO Run scan

	return &resp, nil
}
