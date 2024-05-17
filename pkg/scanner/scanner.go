package scanner

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"time"

	"github.com/leaktk/scanner/pkg/config"
	"github.com/leaktk/scanner/pkg/logger"
)

// Scanner holds the config and state for the scanner processes
type Scanner struct {
	cloneQueue   chan *Request
	responses    chan *Response
	scanQueue    chan *Request
	cloneTimeout time.Duration
	cloneWorkers int
	maxScanDepth int
	scanWorkers  int
}

// NewScanner returns a initialized and listening scanner instance that should
// be closed when it's no longer needed.
func NewScanner(cfg *config.Config) *Scanner {
	scanner := &Scanner{
		cloneQueue:   make(chan *Request, cfg.Scanner.MaxCloneQueueSize),
		responses:    make(chan *Response),
		scanQueue:    make(chan *Request, cfg.Scanner.MaxScanQueueSize),
		cloneTimeout: time.Duration(cfg.Scanner.CloneTimeout) * time.Second,
		cloneWorkers: cfg.Scanner.CloneWorkers,
		maxScanDepth: cfg.Scanner.MaxScanDepth,
		scanWorkers:  cfg.Scanner.ScanWorkers,
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
	for i := 0; i < s.cloneWorkers; i++ {
		go s.listenForCloneRequests()
	}
	// Start scan workers
	for i := 0; i < s.scanWorkers; i++ {
		go s.listenForScanRequests()
	}
}

// Watch the clone queue for requests
func (s *Scanner) listenForCloneRequests() {
	for request := range s.cloneQueue {
		var err error = nil
		// Pick the right cloning method (if we get too many of these it might
		// make sense to abstract that behind some kind of provider interface)
		switch request.Kind {
		case "GitRepo":
			err = s.cloneGitRepo(request)
		default:
			logger.Error("unsupported kind kind=%s id=%s", request.Kind, request.ID)
			continue
		}
		// Stop if there was an error and log it
		if err != nil {
			logger.Error("clone error: %s", err)
			continue
		}

		// Now that it's cloned send it on to the scan queue
		s.scanQueue <- request
	}
}

// Watch the scan queue for requests
func (s *Scanner) listenForScanRequests() {
	for request := range s.scanQueue {
		switch request.Kind {
		case "GitRepo":
			s.scanGitRepo(request)
		default:
			logger.Error("unsupported kind kind=%s id=%s", request.Kind, request.ID)
		}
	}
}

// cloneGitRepo handles clones before handing them off to the scans
func (s *Scanner) cloneGitRepo(request *Request) error {
	options := request.GitRepoOptions()

	if options == nil {
		return errors.New("GitRepoOptions is nil")
	}

	if s.maxScanDepth > 0 && options.Depth > s.maxScanDepth {
		logger.Warning("reducing the scan depth to the max scan depth: %d", s.maxScanDepth)
		options.Depth = s.maxScanDepth
	}

	cloneArgs := []string{"clone"}

	if len(options.Proxy) > 0 {
		cloneArgs = append(cloneArgs, "--config")
		cloneArgs = append(cloneArgs, fmt.Sprintf("http.proxy=%s", options.Proxy))
	}

	if len(options.Since) > 0 {
		cloneArgs = append(cloneArgs, "--shallow-since")
		cloneArgs = append(cloneArgs, options.Since)
	}

	if len(options.Branch) > 0 {
		cloneArgs = append(cloneArgs, "--single-branch")
		cloneArgs = append(cloneArgs, "--branch")
		cloneArgs = append(cloneArgs, options.Branch)
	} else {
		cloneArgs = append(cloneArgs, "--no-single-branch")
	}

	if options.Depth > 0 {
		cloneArgs = append(cloneArgs, "--depth")
		// TODO: Confirm this is an issue still in gitleaks 8
		// Overclone by one commit to avoid issues with scanning the grafted commit
		cloneArgs = append(cloneArgs, fmt.Sprint(options.Depth+1))
	}

	// Include the clone URL
	cloneArgs = append(cloneArgs, request.Resource)

	// TODO determine the repo dir and store that on the request somewhere
	// for the scan. Might make sense to determine that right when the
	// request happens and just reference here. Also if there are any errors
	// the repo dir needs to be cleaned up.

	ctx, cancel := context.WithTimeout(context.Background(), s.cloneTimeout)
	defer cancel()

	gitClone := exec.CommandContext(ctx, "git", cloneArgs...)
	output, err := gitClone.CombinedOutput()

	if err != nil {
		logger.Error("git clone error: %v", err)
		logger.Error("%s", output)
	} else {
		logger.Debug("%s", output)
	}

	if ctx.Err() == context.DeadlineExceeded {
		return fmt.Errorf("clone timeout exceeded id=%s %v", request.ID, ctx.Err())
	}

	return nil
}

// scanGitRepo handles git repo scans
func (s *Scanner) scanGitRepo(request *Request) {
	s.responses <- &Response{
		Request: RequestDetails{
			ID:       request.ID,
			Kind:     request.Kind,
			Resource: request.Resource,
		},
	}
}
