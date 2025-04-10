package cmd

import (
	"fmt"
	"net/http"
	"runtime"
)

// globalUserAgent is the useragent used by default by http
var globalUserAgent = fmt.Sprintf("leaktk-scanner/%s (%s %s)", shortVersion(), runtime.GOOS, runtime.GOARCH)

type userAgentTransport struct {
	rt http.RoundTripper
}

func (uat *userAgentTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.Header.Set("User-Agent", globalUserAgent)
	return uat.rt.RoundTrip(req)
}

func init() {
	http.DefaultTransport = &userAgentTransport{
		rt: http.DefaultTransport,
	}
}

func shortVersion() string {
	if len(Version) > 0 {
		if len(Commit) > 0 {
			return Version + "@" + Commit
		}
		return Version
	}
	return "unknown"
}
