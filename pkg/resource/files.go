package resource

import (
	iofs "io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/leaktk/scanner/pkg/response"

	"github.com/leaktk/scanner/pkg/fs"
	"github.com/leaktk/scanner/pkg/logger"
)

// Files provides a way to scan file systems
type Files struct {
	// Provide common helper functions
	BaseResource
	path    string
	options *FilesOptions
}

// FilesOptions are options for the Files resource
type FilesOptions struct {
	// The scan priority
	Priority int `json:"priority"`
}

// NewFiles returns a configured Files resource for the scanner to scan
func NewFiles(path string, options *FilesOptions) *Files {
	return &Files{
		path:    filepath.Clean(path),
		options: options,
	}
}

// Kind of resource (always returns Files here)
func (r *Files) Kind() string {
	return "Files"
}

// String representation of the resource
func (r *Files) String() string {
	return r.path
}

// Clone the resource to the desired local location and store the path
func (r *Files) Clone(path string) error {
	// no-op
	return nil
}

// ClonePath returns where this repo has been cloned if cloned else ""
func (r *Files) ClonePath() string {
	return r.path
}

// Depth returns the depth for things that have version control
func (r *Files) Depth() uint16 {
	return 0
}

// EnrichResult enriches the result with contextual information
func (r *Files) EnrichResult(result *response.Result) *response.Result {
	result.Kind = response.GeneralResultKind
	return result
}

// SetDepth allows you to adjust the depth for the resource
func (r *Files) SetDepth(depth uint16) {
	// no-op
}

// SetCloneTimeout lets you adjust the timeout before the clone aborts
func (r *Files) SetCloneTimeout(timeout time.Duration) {
	// no-op
}

// Since returns the date after which things should be scanned for things
// that have versions
func (r *Files) Since() string {
	return ""
}

// ReadFile provides a way to access values in the JSON data
func (r *Files) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(filepath.Join(r.path, filepath.Clean(path)))
}

// Walk traverses the JSON data structure like it's a directory tree
func (r *Files) Walk(fn WalkFunc) error {
	// Handle if path is a file
	if fs.FileExists(r.path) {
		file, err := os.Open(r.path)
		if err != nil {
			return err
		}
		defer file.Close()

		// path is empty because it's not in a directory
		return fn("", file)
	}

	return filepath.WalkDir(r.path, func(path string, d iofs.DirEntry, err error) error {
		if err != nil {
			r.Error(logger.ScanError, "could not walk path %w path=%q", err, path)
			return nil
		}

		relPath, err := filepath.Rel(r.path, path)
		if err != nil {
			r.Error(logger.ScanError, "could generate relative path: %w path=%q", err, path)
			return nil
		}

		if d.IsDir() {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return nil
		}

		if info.Mode()&os.ModeSymlink != 0 {
			r.Info(logger.ScanDetail, "skipping symlink path=%q", relPath)
			return nil
		}

		file, err := os.Open(filepath.Clean(path))
		if err != nil {
			r.Error(logger.ScanError, "could not open file: %w path=%q", err, relPath)
			return nil
		}
		defer file.Close()

		return fn(relPath, file)
	})
}

// Priority returns the scan priority
func (r *Files) Priority() int {
	return r.options.Priority
}
