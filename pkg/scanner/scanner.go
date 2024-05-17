package scanner

import (
	"errors"
	"fmt"
	"github.com/leaktk/scanner/pkg/config"

	"github.com/leaktk/scanner/pkg/logger"
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

	switch request.Kind {
	case "GitRepo":
		return ScanGitRepo(cfg, request)
	default:
		return nil, fmt.Errorf("%s is an unsupported kind", request.Kind)
	}
}

// scanGitRepo handles git repo scans
func scanGitRepo(cfg *config.Config, request *Request) (*Response, error) {
	options := request.GitRepoOptions()

	if options == nil {
		return errors.New("GitRepoOptions is nil")
	}

	if cfg.Scanner.MaxScanDepth > 0 && options.Depth > cfg.Scanner.MaxScanDepth {
		logger.Warning("reducing the scan depth to the max scan depth: %d", cfg.Scanner.MaxScanDepth)
		options.Depth = cfg.Scanner.MaxScanDepth
	}
}
