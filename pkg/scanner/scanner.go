package scanner

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/zricethezav/gitleaks/v8/report"

	"github.com/leaktk/leaktk/pkg/config"
	"github.com/leaktk/leaktk/pkg/fs"
	"github.com/leaktk/leaktk/pkg/id"
	"github.com/leaktk/leaktk/pkg/logger"
	"github.com/leaktk/leaktk/pkg/proto"
	"github.com/leaktk/leaktk/pkg/queue"
	"github.com/leaktk/leaktk/pkg/scanner/gitleaks"

	httpclient "github.com/leaktk/leaktk/pkg/http"
)

// Set initial queue size. The queue can grow over time if needed
const queueSize = 1024

// Scanner holds the config and state for the scanner processes
type Scanner struct {
	allowLocal          bool
	cloneQueue          *queue.PriorityQueue[*Request]
	cloneTimeout        time.Duration
	includeResponseLogs bool
	maxArchiveDepth     int
	maxDecodeDepth      int
	maxScanDepth        int
	patterns            *Patterns
	resourceDir         string
	responseQueue       *queue.PriorityQueue[*proto.Response]
	scanQueue           *queue.PriorityQueue[*Request]
	scanWorkers         int
}

// NewScanner returns a initialized and listening scanner instance that should
// be closed when it's no longer needed.
func NewScanner(cfg *config.Config) *Scanner {
	scanner := &Scanner{
		allowLocal:          cfg.Scanner.AllowLocal,
		cloneTimeout:        time.Duration(cfg.Scanner.CloneTimeout) * time.Second,
		clonesDir:           filepath.Join(cfg.Scanner.Workdir, "clones"),
		includeResponseLogs: cfg.Scanner.IncludeResponseLogs,
		maxArchiveDepth:     int(cfg.Scanner.MaxArchiveDepth),
		maxDecodeDepth:      int(cfg.Scanner.MaxDecodeDepth),
		maxScanDepth:        int(cfg.Scanner.MaxScanDepth),
		patterns:            NewPatterns(&cfg.Scanner.Patterns, httpclient.NewClient()),
		responseQueue:       queue.NewPriorityQueue[*proto.Response](queueSize),
		scanQueue:           queue.NewPriorityQueue[*Request](queueSize),
		scanWorkers:         cfg.Scanner.ScanWorkers,
	}

	scanner.start()
	return scanner
}

// Recv sends scan responses to a callback function
func (s *Scanner) Recv(fn func(*proto.Response)) {
	s.responseQueue.Recv(func(msg *queue.Message[*proto.Response]) {
		fn(msg.Value)
	})
}

// Send accepts a request for scanning and puts it in the queues
func (s *Scanner) Send(request *proto.Request) {
	logger.Info("queueing scan: id=%q", request.ID)
	s.scanQueue.Send(&queue.Message[*protoRequest]{
		Priority: (*proto.CommonOptions)(request.Options).Priority,
		Value:    request,
	})
}

// start kicks off the background workers
func (s *Scanner) start() {
	// Start workers
	for i := int(0); i < s.scanWorkers; i++ {
		go s.listen()
	}
}

// Watch the scan queue for requests
func (s *Scanner) listen() {
	s.scanQueue.Recv(func(msg *queue.Message[*proto.Request]) {
		logger.Info("starting scan: id=%q", msg.Value.ID)
		ctx := context.Background()
		request := msg.Value
		detectorOpts := gitleaks.DetectorOpts{
			MaxArchiveDepth: s.maxArchiveDepth,
			MaxDecodeDepth:  s.maxDecodeDepth,
		}

		var clonePath string
		var err error
		if cloneNeeded(request) {
			cloneCtx := ctx
			if s.cloneTimeout > 0 {
				cloneCtx = context.WithTimeout(cloneCtx, s.cloneTimeout)
			}

			clonePath, err = cloneResource(cloneCtx, s.clonesDir, request)
			if err != nil {
				err = fmt.Errorf("clone failed: %w", err)
			}
		} else if isLocalResource(request) && !s.allowLocal {
			err = errors.New("local scans not allowed")
		}

		if err != nil {
			logger.Critical("scan failed: %v id=%q", err, request.ID)
			logger.Info("queueing response: id=%q", request.ID)
			s.responseQueue.Send(&queue.Message[*proto.Response]{
				Priority: msg.Priority,
				Value: &proto.Response{
					ID:        id.ID(),
					RequestID: request.ID,
				},
			})
			return
		}

		scanPath := resourceScanPath(request, clonePath)
		if len(scanPath) > 0 && fs.PathExists(scanPath) && !fs.FileExists(scanPath) {
			detectorOpts.SourcePath = scanPath
			rawAdditionalConfig, err := os.ReadFile(filepath.Join(scanPath, ".gitleaks.toml"))
			if err == nil {
				detectorOpts.AdditionalConfig = string(rawClonedConfig)
			}
			baselinePath := filepath.Join(scanPath, ".gitleaksbaseline")
			if fs.FileExists(baselinePath) {
				detectorOpts.BaselinePath = baselinePath
			}
			ignorePath := filepath.Join(scanPath, ".gitleaksignore")
			if fs.FileExists(ignorePath) {
				detectorOpts.IgnorePath = ignorePath
			}
		}

		detector, err := gitleaks.NewDetector(s.patterns.Gitleaks(), detectorOpts)
		if err != nil {
			logger.Critical("scan failed: could not create detector: %s id=%q", err, request.ID)
			logger.Info("queueing response: id=%q", request.ID)
			s.responseQueue.Send(&queue.Message[*proto.Response]{
				Priority: msg.Priority,
				Value: &proto.Response{
					ID:        id.ID(),
					RequestID: request.ID,
				},
			})
			return
		}

		var findings []report.Finding
		switch opts := request.Options.(type) {
		case *proto.GitRepoOptions:
			findings, err := gitleaks.ScanGit(ctx, detector, scanPath, gitleaks.GitScanOpts{
				Branch:   opts.Branch,
				Depth:    opts.Depth,
				Since:    opts.Since,
				Staged:   opts.Staged,
				Unstaged: opts.Unstaged,
			})
		case *proto.JSONDataOptions:
			findings, err := gitleaks.ScanJSON(ctx, detector, request.Resource, gitleaks.JSONScanOpts{
				FetchURLPatterns: strings.Split(opts.FetchURLs, ":"),
			})
		default:
			if request.Kind == proto.TextRequestKind {
				findings, err := gitleaks.ScanReader(ctx, detector, strings.NewStringReader(request.Resource))
			} else {
				findings, err := gitleaks.ScanFiles(ctx, detector, scanPath)
			}
		}

		results = make([]*proto.Result, findings)
		for i, finding := range findings {
			results[i] = findingToResult(request, &finding)
		}

		if len(clonePath) > 0 {
			if err := os.RemoveAll(clonePath); err != nil {
				logger.Error("could not remove clone path: %v path=%q id=%q", err, clonePath, request.ID)
			}
		}

		logger.Info("queueing response: id=%q", request.ID)
		s.responseQueue.Send(&queue.Message[*proto.Response]{
			Priority: msg.Priority,
			Value: &proto.Response{
				// TODO: pre 1.0 consier making this be the request id and maybe
				// making this whole thing a little more like a JSON-RPC
				// response (i.e. id, result & error fields that follow the
				// spec)
				ID:        id.ID(),
				Results:   results,
				RequestID: request.ID,
			},
		})
	})
}

func findingToResult(request proto.Request, finding *report.Finding) *proto.Result {
	result := &proto.Result{
		ID: id.ID(
			request.Resource,
			finding.Commit,
			finding.File,
			strconv.Itoa(finding.StartLine),
			strconv.Itoa(finding.StartColumn),
			strconv.Itoa(finding.EndLine),
			strconv.Itoa(finding.EndColumn),
			finding.RuleID,
		),
		Secret:  finding.Secret,
		Match:   finding.Match,
		Context: finding.Line,
		Entropy: finding.Entropy,
		Date:    finding.Date,
		Notes:   map[string]string{},
		Contact: proto.Contact{
			Name:  finding.Author,
			Email: finding.Email,
		},
		Rule: proto.Rule{
			ID:          finding.RuleID,
			Description: finding.Description,
			// TODO: pre 1.0 tags should be moved up to result since
			// tags can be dynamic
			Tags: finding.Tags,
		},
		Location: proto.Location{
			Version: finding.Commit,
			Path:    finding.File,
			Start: proto.Point{
				Line:   finding.StartLine,
				Column: finding.StartColumn,
			},
			End: proto.Point{
				Line:   finding.EndLine,
				Column: finding.EndColumn,
			},
		},
	}

	switch request.Kind {
	case proto.GitRepoRequestKind:
		notes["gitleaks_fingerprint"] = finding.Fingerprint
		notes["message"] = finding.Message
		result.Kind = proto.GitCommitResultKind
	// case proto.ContainerImageRequestKind:
	// case proto.FilesImageRequestKind:
	// case proto.URLRequestKind:
	// TODO: add more here and other handlers for the different kinds
	default:
		result.Kind = proto.GenericResultKind
	}

	return result
}
