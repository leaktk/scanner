package scanner

import (
	"github.com/leaktk/scanner/pkg/config"
)

// Scan takes a recquest scans the resource and returns the results response
func Scan(cfg *config.Config, request *Request) (*Response, error) {
	response := Response{
		Request: RequestDetails{
			ID:       request.ID,
			Kind:     request.Kind,
			Resource: request.Resource,
		},
	}

	// TODO Run scan

	return &response, nil
}
