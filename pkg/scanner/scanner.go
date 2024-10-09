package scanner

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/leaktk/scanner/pkg/config"
	"github.com/leaktk/scanner/pkg/fs"
	"github.com/leaktk/scanner/pkg/id"
	"github.com/leaktk/scanner/pkg/logger"
	"github.com/leaktk/scanner/pkg/queue"
	"github.com/leaktk/scanner/pkg/resource"
	"github.com/leaktk/scanner/pkg/response"
)

// Set initial queue size. The queue can grow over time if needed
const queueSize = 1024

// Scanner holds the config and state for the scanner processes
type Scanner struct {
	backends      []Backend
	cloneQueue    *queue.PriorityQueue[*Request]
	cloneTimeout  time.Duration
	cloneWorkers  uint16
	maxScanDepth  uint16
	resourceDir   string
	responseQueue *queue.PriorityQueue[*response.Response]
	scanQueue     *queue.PriorityQueue[*Request]
	scanWorkers   uint16
}

// NewScanner returns a initialized and listening scanner instance that should
// be closed when it's no longer needed.
func NewScanner(cfg *config.Config) *Scanner {
	scanner := &Scanner{
		cloneQueue:    queue.NewPriorityQueue[*Request](queueSize),
		cloneTimeout:  time.Duration(cfg.Scanner.CloneTimeout) * time.Second,
		cloneWorkers:  cfg.Scanner.CloneWorkers,
		maxScanDepth:  cfg.Scanner.MaxScanDepth,
		resourceDir:   filepath.Join(cfg.Scanner.Workdir, "resources"),
		responseQueue: queue.NewPriorityQueue[*response.Response](queueSize),
		scanQueue:     queue.NewPriorityQueue[*Request](queueSize),
		scanWorkers:   cfg.Scanner.ScanWorkers,
		backends: []Backend{
			NewGitleaks(NewPatterns(&cfg.Scanner.Patterns, &http.Client{})),
		},
	}

	scanner.start()
	return scanner
}

// Recv sends scan responses to a callback function
func (s *Scanner) Recv(fn func(*response.Response)) {
	s.responseQueue.Recv(func(msg *queue.Message[*response.Response]) {
		fn(msg.Value)
	})
}

// Send accepts a request for scanning and puts it in the queues
func (s *Scanner) Send(request *Request) {
	logger.Debug("queueing clone: request_id=%q", request.ID)
	s.cloneQueue.Send(&queue.Message[*Request]{
		Priority: request.Priority(),
		Value:    request,
	})
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
	s.cloneQueue.Recv(func(msg *queue.Message[*Request]) {
		request := msg.Value
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
				request.Errors = append(request.Errors, response.LeakTKError{
					Fatal:   true,
					Code:    response.CloneError,
					Message: err.Error(),
				})
			}
		}

		// Now that it's cloned send it on to the scan queue
		logger.Debug("queueing scan: request_id=%q", request.ID)
		s.scanQueue.Send(msg)
	})
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
	s.scanQueue.Recv(func(msg *queue.Message[*Request]) {
		request := msg.Value
		reqResource := request.Resource

		results := make([]*response.Result, 0)

		if fs.PathExists(reqResource.ClonePath()) {
			for _, backend := range s.backends {
				logger.Info("starting scan: request_id=%q scanner_backend=%q", request.ID, backend.Name())

				backendResults, err := backend.Scan(reqResource)
				if err != nil {
					logger.Error("scan error: request_id=%q error=%q", request.ID, err.Error())
					request.Errors = append(request.Errors, response.LeakTKError{
						Fatal:   true,
						Code:    response.ScanError,
						Message: err.Error(),
					})
				}
				if backendResults != nil {
					results = append(results, backendResults...)
				}
			}
			if err := s.removeResourceFiles(reqResource); err != nil {
				logger.Error("resource file cleanup error: request_id=%q error=%q", request.ID, err.Error())
				request.Errors = append(request.Errors, response.LeakTKError{
					Fatal:   false,
					Code:    response.ResourceCleanupError,
					Message: err.Error(),
				})
			}
		} else {
			logger.Error("skipping scan due to missing clone path: request_id=%q", request.ID)
			request.Errors = append(request.Errors, response.LeakTKError{
				Fatal:   true,
				Code:    response.CloneError,
				Message: fmt.Sprintf("missing clone path: request_id=%q (%s)", request.ID, reqResource.ClonePath()),
			})
		}

		s.responseQueue.Send(&queue.Message[*response.Response]{
			Priority: msg.Priority,
			Value: &response.Response{
				ID:        id.ID(),
				Results:   results,
				Errors:    request.Errors,
				RequestID: request.ID,
			},
		})
	})
}
