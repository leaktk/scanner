package http

import (
	"net/http"

	"github.com/leaktk/scanner/version"
)

// HTTPClient provides an interface for working with Go's http client or
// swapping it out with other types for testing
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// NewClient creates a http client with preferred configuration
func NewClient() *http.Client {
	return &http.Client{
		Transport: &RoundTripper{
			rt: http.DefaultTransport,
		},
	}
}

type RoundTripper struct {
	rt http.RoundTripper
}

func (rt *RoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.Header.Set("User-Agent", version.UserAgent())
	return rt.rt.RoundTrip(req)
}
