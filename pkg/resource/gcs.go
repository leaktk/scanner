package resource

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"sync"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"

	"github.com/leaktk/scanner/pkg/logger"
	"github.com/leaktk/scanner/pkg/response"
)

var (
    getClientMu sync.Mutex
    storageClient *storage.Client
)

// newStorageClient returns a usable storage client for resources. Since
// storage.Client is safe across multiple threads, it returns the same
// client each time
func newStorageClient() (*storage.Client, error) {
	// avoid race conditions
	getClientMu.Lock()
	defer getClientMu.Unlock()

	// see if it has already been created
	if storageClient != nil {
		return storageClient, nil
	}

	// set a new global one
	var err error
	storageClient, err = storage.NewClient(context.Background())
	return storageClient, err
}

// GCS provides a way to scan Google Cloud Storage buckets and objects
type GCS struct {
	// Provide common helper functions
	BaseResource
	url          string
	path         string
	bucket       *storage.BucketHandle
	options      *GCSOptions
	cloneTimeout time.Duration
	clonePath    string
}

// GCSOptions are options for the GCS resource
type GCSOptions struct {
	// The scan priority
	Priority int `json:"priority"`
	// The google cloud storage query
	Query *storage.Query `json:"query"`
}

// NewGCS returns a configured Google Cloud Storage resource
func NewGCS(url string, options *GCSOptions) *GCS {
	return &GCS{
		url:     url,
		options: options,
	}
}

// Kind of resource (always returns GCS here)
func (r *GCS) Kind() string {
	return "GCS"
}

// String representation of the resource
func (r *GCS) String() string {
	return r.url
}

// Clone the resource to the desired local location and store the path
func (r *GCS) Clone(path string) error {
	// Let the scanner know that this resource was cloned
	// This can also be used as a cache if needed
	r.clonePath = path
	if err := os.MkdirAll(r.clonePath, 0700); err != nil {
		return err
	}

	// Parse the GCS url into usable parts
	u, err := url.Parse(r.url)
	if err != nil {
		return fmt.Errorf("could not parse GCS url: %w url=%q", err, r.url)
	}

	// Get our cloud storage client
	client, err := newStorageClient()
	if err != nil {
		return fmt.Errorf("could not create client: %w", err)
	}

	// Set up the bucket and other details for walking the object
	if len(u.Path) > 0 {
		r.path = u.Path[1:] // ignore the first slash
	}
	r.bucket = client.Bucket(u.Host)

	return nil
}

// ClonePath returns where this repo has been cloned if cloned else ""
func (r *GCS) ClonePath() string {
	return r.clonePath
}

// Depth returns the depth for things that have version control
func (r *GCS) Depth() uint16 {
	return 0
}

// EnrichResult enriches the result with contextual information
func (r *GCS) EnrichResult(result *response.Result) *response.Result {
	result.Kind = response.GeneralResultKind
	return result
}

// SetDepth allows you to adjust the depth for the resource
func (r *GCS) SetDepth(depth uint16) {
	// no-op
}

// SetCloneTimeout lets you adjust the timeout before the clone aborts
func (r *GCS) SetCloneTimeout(timeout time.Duration) {
	r.cloneTimeout = timeout
}

// Since returns the date after which things should be scanned for things
// that have versions
func (r *GCS) Since() string {
	return ""
}

// contextWithCancel returns a context configured for this timeout
func (r *GCS) contextWithCancel() (context.Context, context.CancelFunc) {
	if r.cloneTimeout > 0 {
		return context.WithTimeout(context.Background(), r.cloneTimeout)
	}
	return context.Background(), func() {}
}

// objectReader returns a reader with the proper context configured
func (r *GCS) objectReader(path string) (io.Reader, context.CancelFunc, error) {
	ctx, cancel := r.contextWithCancel()
	reader, err := r.bucket.Object(path).NewReader(ctx)
	if err != nil {
		return nil, cancel, fmt.Errorf(
			"could not read object: %w bucket=%q object=%q",
			err, r.bucket.BucketName(), path,
		)
	}

	return reader, cancel, nil
}

// ReadFile provides a way to access values in the JSON data
func (r *GCS) ReadFile(path string) ([]byte, error) {
	reader, cancel, err := r.objectReader(path)
	defer cancel()
	if err != nil {
		return nil, err
	}
	return io.ReadAll(reader)
}

// Walk traverses the JSON data structure like it's a directory tree
func (r *GCS) Walk(fn WalkFunc) error {
	if r.options.Query == nil {
		reader, cancel, err := r.objectReader(r.path)
		defer cancel()
		if err != nil {
			return err
		}
		return fn(r.path, reader)
	}

	// Iterate through the objects based on a query
	// TODO: get the clone timeout worked out correctly for all of this
	ctx, cancel := r.contextWithCancel()
	defer cancel()
	objects := r.bucket.Objects(ctx, r.options.Query)
	for {
		objectAttrs, err := objects.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			r.Error(
				logger.ScanError,
				"could not list objects: %w bucket=%q",
				err, r.bucket.BucketName(),
			)
			continue
		}

		path := objectAttrs.Name
		reader, cancel, err := r.objectReader(path)
		if err != nil {
			r.Error(logger.ScanError, "%w", err)
			defer cancel()
			continue
		}

		err = fn(path, reader)
		cancel()
		if err != nil {
			return err
		}
	}

	return nil
}

// Priority returns the scan priority
func (r *GCS) Priority() int {
	return r.options.Priority
}
