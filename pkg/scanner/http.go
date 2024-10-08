package scanner

import (
	"net/http"
)

// HTTPClient provides an interface for working with Go's http client or
// swapping it out with other types for testing
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}
