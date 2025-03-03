package scanner

import (
	"os"
	"path/filepath"
	"time"

	"github.com/leaktk/scanner/pkg/config"
	"github.com/leaktk/scanner/pkg/fs"
	"github.com/leaktk/scanner/pkg/http"
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
	allowLocal          bool
	backends            []Backend
	cloneQueue          *queue.PriorityQueue[*Request]
	cloneTimeout        time.Duration
	cloneWorkers        uint16
	includeResponseLogs bool
	maxScanDepth        uint16
	resourceDir         string
	responseQueue       *queue.PriorityQueue[*response.Response]
	scanQueue           *queue.PriorityQueue[*Request]
	scanWorkers         uint16
}

// NewScanner returns a initialized and listening scanner instance that should
// be closed when it's no longer needed.
func NewScanner(cfg *config.Config) *Scanner {
	scanner := &Scanner{
		allowLocal:          cfg.Scanner.AllowLocal,
		cloneQueue:          queue.NewPriorityQueue[*Request](queueSize),
		cloneTimeout:        time.Duration(cfg.Scanner.CloneTimeout) * time.Second,
		cloneWorkers:        cfg.Scanner.CloneWorkers,
		maxScanDepth:        cfg.Scanner.MaxScanDepth,
		resourceDir:         filepath.Join(cfg.Scanner.Workdir, "resources"),
		responseQueue:       queue.NewPriorityQueue[*response.Response](queueSize),
		scanQueue:           queue.NewPriorityQueue[*Request](queueSize),
		scanWorkers:         cfg.Scanner.ScanWorkers,
		includeResponseLogs: cfg.Scanner.IncludeResponseLogs,
		backends: []Backend{
			NewGitleaks(
				cfg.Scanner.MaxDecodeDepth,
				NewPatterns(&cfg.Scanner.Patterns, http.NewClient()),
			),
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
	logger.Info("queueing clone: request_id=%q resource_id=%q", request.ID, request.Resource.ID())
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
	// This ensures that things waiting on responses can mark them as done
	s.cloneQueue.Recv(func(msg *queue.Message[*Request]) {
		request := msg.Value
		reqResource := request.Resource
		reqResource.IncludeLogs(s.includeResponseLogs)

		if request.Resource.IsLocal() && !s.allowLocal {
			reqResource.Error(logger.LocalScanDisabled, "local resources not allowed: request_id=%q", request.ID)
			s.responseQueue.Send(&queue.Message[*response.Response]{
				Priority: msg.Priority,
				Value: &response.Response{
					ID:        id.ID(),
					Results:   make([]*response.Result, 0),
					Logs:      reqResource.Logs(),
					RequestID: request.ID,
				},
			})
			return
		}

		if s.cloneTimeout > 0 {
			logger.Debug("setting clone timeout: request_id=%q resource_id=%q timeout=%v", request.ID, reqResource.ID(), s.cloneTimeout.Seconds())
			reqResource.SetCloneTimeout(s.cloneTimeout)
		}

		if s.maxScanDepth > 0 && reqResource.Depth() > s.maxScanDepth {
			logger.Warning("reducing scan depth: request_id=%q resource_id=%q old_depth=%v new_depth=%v", request.ID, reqResource.ID(), reqResource.Depth(), s.maxScanDepth)
			reqResource.SetDepth(s.maxScanDepth)
		}

		if reqResource.Path() == "" {
			logger.Info("starting clone: request_id=%q resource_id=%q", request.ID, reqResource.ID())
			if err := reqResource.Clone(s.resourceClonePath(reqResource)); err != nil {
				reqResource.Critical(logger.CloneError, "clone error: request_id=%q error=%q", request.ID, err.Error())
			}
		}

		// Now that it's cloned send it on to the scan queue
		logger.Info("queueing scan: request_id=%q resource_id=%q", request.ID, reqResource.ID())
		s.scanQueue.Send(msg)
	})
}

func (s *Scanner) resourceFilesPath(reqResource resource.Resource) string {
	return filepath.Join(s.resourceDir, reqResource.ID())
}

func (s *Scanner) resourceClonePath(reqResource resource.Resource) string {
	return filepath.Join(s.resourceFilesPath(reqResource), "clone")
}

// removeResourceFiles clears out any left over resource files for scan
func (s *Scanner) removeResourceFiles(reqResource resource.Resource) error {
	return os.RemoveAll(s.resourceFilesPath(reqResource))
}

// Watch the scan queue for requests
func (s *Scanner) listenForScanRequests() {
	s.scanQueue.Recv(func(msg *queue.Message[*Request]) {
		request := msg.Value
		reqResource := request.Resource

		results := make([]*response.Result, 0)

		if fs.PathExists(reqResource.Path()) {
			for _, backend := range s.backends {
				logger.Info("starting scan: request_id=%q resource_id=%q scanner_backend=%q", request.ID, reqResource.ID(), backend.Name())

				backendResults, err := backend.Scan(reqResource)
				if err != nil {
					reqResource.Critical(logger.ScanError, "scan error: request_id=%q error=%q", request.ID, err.Error())
				}
				if backendResults != nil {
					results = append(results, backendResults...)
				}
			}
			if err := s.removeResourceFiles(reqResource); err != nil {
				reqResource.Error(logger.ResourceCleanupError, "resource file cleanup error: request_id=%q error=%q", request.ID, err.Error())
			}
		} else {
			reqResource.Critical(logger.ScanError, "skipping scan due to missing clone path: request_id=%q", request.ID)

		}

		logger.Info("queueing response: request_id=%q resource_id=%q", request.ID, reqResource.ID())
		s.responseQueue.Send(&queue.Message[*response.Response]{
			Priority: msg.Priority,
			Value: &response.Response{
				ID:        id.ID(),
				Results:   results,
				Logs:      reqResource.Logs(),
				RequestID: request.ID,
			},
		})
	})
}
