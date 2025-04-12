package resource

import (
	"io"
	iofs "io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/mholt/archives"

	"github.com/leaktk/scanner/pkg/logger"
	"github.com/leaktk/scanner/pkg/response"
)

// Files provides a way to scan file systems
type Files struct {
	// Provide common helper functions
	BaseResource
	fs      iofs.FS
	root    string
	options *FilesOptions
}

// FilesOptions are options for the Files resource
type FilesOptions struct {
	// The scan priority
	Priority int `json:"priority"`
}

// NewFiles returns a configured Files resource for the scanner to scan
func NewFiles(root string, options *FilesOptions) *Files {
	root = filepath.Clean(root)

	return &Files{
		fs:      &archives.DeepFS{Root: root},
		root:    root,
		options: options,
	}
}

// Kind of resource (always returns Files here)
func (r *Files) Kind() string {
	return "Files"
}

// String representation of the resource
func (r *Files) String() string {
	return r.root
}

// Clone the resource to the desired local location and store the path
func (r *Files) Clone(path string) error {
	// no-op
	return nil
}

// ClonePath returns where this repo has been cloned if cloned else ""
func (r *Files) ClonePath() string {
	return r.root
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
	// Clean the path before passing it to the fs. This also turns "" -> "."
	// which is important given we don't use "." to represent an empty path
	file, err := r.fs.Open(filepath.Clean(path))
	if err != nil {
		return nil, err
	}
	return io.ReadAll(file)
}

// Walk traverses the JSON data structure like it's a directory tree
func (r *Files) Walk(fn WalkFunc) error {
	return iofs.WalkDir(r.fs, ".", func(path string, d iofs.DirEntry, err error) error {
		if path == "." {
			path = ""
		}

		if err != nil {
			r.Error(logger.ScanError, "could not walk path: path=%q error=%q", path, err)
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
			r.Info(logger.ScanDetail, "skipping symlink: path=%q", path)
			return nil
		}

		// Clean the path before passing it to the fs. This also turns "" -> "."
		// which is important given we don't use "." to represent an empty path
		file, err := r.fs.Open(filepath.Clean(path))
		if err != nil {
			r.Error(logger.ScanError, "could not open file: path=%q error=%q", path, err)
			return nil
		}
		defer file.Close()

		// Let the calling code know if we've entered into a sub resource
		if dfs, ok := r.fs.(*archives.DeepFS); ok {
			if outer, inner := dfs.SplitPath(path); len(inner) != 0 {
				path = outer + ":" + inner
			}
		}

		return fn(path, file)
	})
}

// Priority returns the scan priority
func (r *Files) Priority() int {
	return r.options.Priority
}
