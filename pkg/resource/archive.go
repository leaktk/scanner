package resource

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/mholt/archives"

	fsutil "github.com/leaktk/scanner/pkg/fs"
	"github.com/leaktk/scanner/pkg/logger"
	"github.com/leaktk/scanner/pkg/response"
)

// Archive provides a way to interact with archive files
type Archive struct {
	// Provide common helper functions
	BaseResource
	path    string
	stream  archives.ReaderAtSeeker
	options *ArchiveOptions
}

// ArchiveOptions are options for the Archive resource
type ArchiveOptions struct {
	// The scan priority
	Priority int `json:"priority"`
}

func NewArchive(path string, options *ArchiveOptions) *Archive {
	return &Archive{
		path:    filepath.Clean(path),
		options: options,
	}
}

// Kind of resource (always returns Archive here)
func (r *Archive) Kind() string {
	return "Archive"
}

// String representation of the resource
func (r *Archive) String() string {
	return r.path
}

// Clone the resource to the desired local location and store the path
func (r *Archive) Clone(path string) error {
	return nil
}

// ClonePath returns where this repo has been cloned if cloned else ""
func (r *Archive) ClonePath() string {
	return r.path
}

// Depth returns the depth for things that have version control
func (r *Archive) Depth() uint16 {
	return 0
}

// EnrichResult enriches the result with contextual information
func (r *Archive) EnrichResult(result *response.Result) *response.Result {
	result.Kind = response.GeneralResultKind
	return result
}

// SetDepth allows you to adjust the depth for the resource
func (r *Archive) SetDepth(depth uint16) {
	// no-op
}

// SetCloneTimeout lets you adjust the timeout before the clone aborts
func (r *Archive) SetCloneTimeout(timeout time.Duration) {
	// no-op
}

// Since returns the date after which things should be scanned for things
// that have versions
func (r *Archive) Since() string {
	return ""
}

func (r *Archive) withFileSystem(fn func(fs.FS) error) error {
	if r.stream == nil {
		// Handle if path is a file
		if fsutil.FileExists(r.path) {
			file, err := os.Open(r.path)
			if err != nil {
				return err
			}

			defer file.Close()
			r.stream = file
		}
	}

	ctx := context.Background()
	fsys, err := archives.FileSystem(ctx, r.path, r.stream)
	if err != nil {
		return nil
	}

	return fn(fsys)
}

// ReadFile provides a way to access values in the archive
func (r *Archive) ReadFile(path string) ([]byte, error) {
	var data []byte

	err := r.withFileSystem(func(fsys fs.FS) error {
		file, err := fsys.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		data, err = io.ReadAll(file)
		return err
	})

	return data, err
}

// Walk the archive
func (r *Archive) Walk(fn WalkFunc) error {
	return r.withFileSystem(func(fsys fs.FS) error {
		return fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
			deepPath := fmt.Sprintf("%s:%s", r.path, path)

			if err != nil {
				r.Error(logger.ScanError, "could not walk path: path=%q error=%q", deepPath, err)
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
				r.Info(logger.ScanDetail, "skipping symlink: path=%q", deepPath)
				return nil
			}

			file, err := fsys.Open(filepath.Clean(path))
			if err != nil {
				r.Error(logger.ScanError, "could not open file: path=%q error=%q", deepPath, err)
				return nil
			}
			defer file.Close()

			return fn(deepPath, file)
		})
	})
}

// Priority returns the scan priority
func (r *Archive) Priority() int {
	return r.options.Priority
}
