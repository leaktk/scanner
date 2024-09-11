package scanner

import (
	"github.com/leaktk/scanner/pkg/resource"
	"github.com/leaktk/scanner/pkg/response"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/leaktk/scanner/pkg/config"
	"github.com/leaktk/scanner/pkg/fs"
	"github.com/leaktk/scanner/pkg/id"
	"github.com/leaktk/scanner/pkg/logger"
)

// Scanner holds the config and state for the scanner processes
type Scanner struct {
	backends     []Backend
	cloneQueue   chan *Request
	cloneTimeout time.Duration
	cloneWorkers uint16
	maxScanDepth uint16
	resourceDir  string
	responses    chan *response.Response
	scanQueue    chan *Request
	scanWorkers  uint16
}

// NewScanner returns a initialized and listening scanner instance that should
// be closed when it's no longer needed.
func NewScanner(cfg *config.Config) *Scanner {
	scanner := &Scanner{
		cloneQueue:   make(chan *Request, cfg.Scanner.MaxCloneQueueSize),
		cloneTimeout: time.Duration(cfg.Scanner.CloneTimeout) * time.Second,
		cloneWorkers: cfg.Scanner.CloneWorkers,
		maxScanDepth: cfg.Scanner.MaxScanDepth,
		resourceDir:  filepath.Join(cfg.Scanner.Workdir, "resources"),
		responses:    make(chan *response.Response),
		scanQueue:    make(chan *Request, cfg.Scanner.MaxScanQueueSize),
		scanWorkers:  cfg.Scanner.ScanWorkers,
		backends: []Backend{
			NewGitleaks(NewPatterns(&cfg.Scanner.Patterns, &http.Client{})),
		},
	}

	scanner.start()
	return scanner
}

// Close closes out all of the queues (make sure to call this)
func (s *Scanner) Close() error {
	close(s.cloneQueue)
	close(s.responses)
	close(s.scanQueue)

	return nil
}

// Responses returns a channel that can be used for subscribing to respones
func (s *Scanner) Responses() <-chan *response.Response {
	return s.responses
}

// Send accepts a request for scanning and puts it in the queues
func (s *Scanner) Send(request *Request) {
	logger.Debug("queueing clone: request_id=%q", request.ID)
	s.cloneQueue <- request
}

// start kicks off the background workers
func (s *Scanner) start() {
	// Start clone workers
	for i := uint16(0); i < s.cloneWorkers; i++ {
		go s.listenForCloneRequests()
	}
	// Start scan workers
	for i := uint16(0); i < s.scanWorkers; i++ {
		go s.listenForScanRequests()
	}
}

// Watch the clone queue for requests
func (s *Scanner) listenForCloneRequests() {
	// This should always send things to the scan queue, even if the clone fails.
	// This ensures that things waiting on respones can mark them as done
	for request := range s.cloneQueue {
		reqResource := request.Resource

		if s.cloneTimeout > 0 {
			logger.Debug("setting clone timeout: request_id=%q timeout=%v", request.ID, s.cloneTimeout.Seconds())
			reqResource.SetCloneTimeout(s.cloneTimeout)
		}

		if s.maxScanDepth > 0 && reqResource.Depth() > s.maxScanDepth {
			logger.Warning("reducing scan depth: request_id=%q old_depth=%v new_depth=%v", request.ID, reqResource.Depth(), s.maxScanDepth)
			reqResource.SetDepth(s.maxScanDepth)
		}

		if reqResource.ClonePath() == "" {
			logger.Info("starting clone: request_id=%q", request.ID)
			if err := reqResource.Clone(s.resourceClonePath(reqResource)); err != nil {
				logger.Error("clone error: request_id=%q error=%q", request.ID, err.Error())
			}
		}

		// Now that it's cloned send it on to the scan queue
		logger.Debug("queueing scan: request_id=%q", request.ID)
		s.scanQueue <- request
	}
}

func (s *Scanner) resourceFilesPath(reqResource resource.Resource) string {
	return filepath.Join(s.resourceDir, reqResource.ID())
}

func (s *Scanner) resourceClonePath(reqResource resource.Resource) string {
	return filepath.Join(s.resourceFilesPath(reqResource), "clone")
}

// removeResourceFiles cleares out any left over resource files for scan
func (s *Scanner) removeResourceFiles(reqResource resource.Resource) error {
	return os.RemoveAll(s.resourceFilesPath(reqResource))
}

// Watch the scan queue for requests
func (s *Scanner) listenForScanRequests() {
	for request := range s.scanQueue {
		reqResource := request.Resource

		results := make([]*response.Result, 0)

		if fs.PathExists(reqResource.ClonePath()) {
			for _, backend := range s.backends {
				logger.Info("starting scan: request_id=%q scanner_backend=%q", request.ID, backend.Name())

				backendResults, err := backend.Scan(reqResource)
				if err != nil {
					logger.Error("scan error: request_id=%q error=%q", request.ID, err.Error())
				}
				if backendResults != nil {
					results = append(results, backendResults...)
				}
			}
			if err := s.removeResourceFiles(reqResource); err != nil {
				logger.Error("resource file cleanup error: request_id=%q error=%q", request.ID, err.Error())
			}
		} else {
			logger.Error("skipping scan due to missing clone path: request_id=%q", request.ID)
		}

		s.responses <- &response.Response{
			ID:      id.ID(),
			Results: results,
			Request: response.RequestDetails{
				ID:       request.ID,
				Kind:     request.Resource.Kind(),
				Resource: request.Resource.String(),
			},
		}
	}
}
