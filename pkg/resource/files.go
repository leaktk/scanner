package resource

import (
	iofs "io/fs"
	"os"
	"path/filepath"
	"time"

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
	// Currently none needed but here for future cases
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
		data, err := r.ReadFile("")
		if err != nil {
			return err
		}

		// path is empty because it's not in a directory
		return fn("", data)
	}

	return filepath.WalkDir(r.path, func(path string, d iofs.DirEntry, err error) error {
		if err != nil {
			logger.Error("could not walk path: path=%q error=%q", path, err)
			return nil
		}

		if d.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(r.path, path)
		if err != nil {
			logger.Error("could generate relative path: path=%q error=%q", path, err)
			return nil
		}

		data, err := r.ReadFile(relPath)
		if err != nil {
			logger.Error("could not read file: path=%q error=%q", path, err)
			return nil
		}

		return fn(relPath, data)
	})
}
