package scanner

import (
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"

	"github.com/leaktk/scanner/pkg/config"
	"github.com/leaktk/scanner/pkg/logger"
	"github.com/leaktk/scanner/pkg/resource"
)

// Scanner holds the config and state for the scanner processes
type Scanner struct {
	cloneQueue   chan *Request
	cloneTimeout time.Duration
	cloneWorkers uint16
	maxScanDepth uint16
	responses    chan *Response
	scanQueue    chan *Request
	scanWorkers  uint16
	resourceDir  string
	backends     []Backend
}

// NewScanner returns a initialized and listening scanner instance that should
// be closed when it's no longer needed.
func NewScanner(cfg *config.Config) *Scanner {
	scanner := &Scanner{
		cloneQueue:   make(chan *Request, cfg.Scanner.MaxCloneQueueSize),
		cloneTimeout: time.Duration(cfg.Scanner.CloneTimeout) * time.Second,
		cloneWorkers: cfg.Scanner.CloneWorkers,
		maxScanDepth: cfg.Scanner.MaxScanDepth,
		responses:    make(chan *Response),
		scanQueue:    make(chan *Request, cfg.Scanner.MaxScanQueueSize),
		scanWorkers:  cfg.Scanner.ScanWorkers,
		resourceDir:  filepath.Join(cfg.Scanner.Workdir, "resources"),
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
func (s *Scanner) Responses() <-chan *Response {
	return s.responses
}

// Send accepts a request for scanning and puts it in the queues
func (s *Scanner) Send(request *Request) {
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
	for request := range s.cloneQueue {
		reqResource := request.Resource

		if s.cloneTimeout > 0 {
			reqResource.SetCloneTimeout(s.cloneTimeout)
		}

		if s.maxScanDepth > 0 && reqResource.Depth() > s.maxScanDepth {
			reqResource.SetDepth(s.maxScanDepth)
			logger.Warning("reduced scan depth: resource_id=%q depth=%d", reqResource.ID(), s.maxScanDepth)
		}

		if err := reqResource.Clone(s.resourceClonePath(reqResource)); err != nil {
			logger.Error("clone error: resource_id=%q error=%q", reqResource.ID(), err.Error())

			if err := s.removeResourceFiles(reqResource); err != nil {
				logger.Error("resource file cleanup error: resource_id=%q error=%q", reqResource.ID(), err.Error())
			}

			continue
		}

		// Now that it's cloned send it on to the scan queue
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

		logger.Warning("TODO: scan %s", reqResource.String())
		var results []*Result

		for _, backend := range s.backends {
			backendResults, err := backend.Scan(reqResource)

			if err != nil {
				logger.Error("scan error: resource_id=%q error=%q", reqResource.ID(), err.Error())
			}

			if backendResults != nil {
				results = append(results, backendResults...)
			}
		}

		if err := s.removeResourceFiles(reqResource); err != nil {
			logger.Error("resource file cleanup error: resource_id=%q error=%q", reqResource.ID(), err.Error())
		}

		s.responses <- &Response{
			ID:      uuid.New().String(),
			Results: results,
			Request: RequestDetails{
				ID:       request.ID,
				Kind:     request.Resource.Kind(),
				Resource: request.Resource.String(),
			},
		}
	}
}
