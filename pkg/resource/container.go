package resource

import (
	"context"
	"fmt"
	"github.com/containers/image/v5/manifest"
	"github.com/containers/image/v5/pkg/blobinfocache"
	"github.com/leaktk/scanner/pkg/fs"
	"github.com/leaktk/scanner/pkg/logger"
	"io"
	iofs "io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/containers/image/v5/docker"
	"github.com/containers/image/v5/types"
)

// Container provides a pull and hold a container
type Container struct {
	// Provide common helper functions
	BaseResource
	clonePath string
	resource  Resource
	location  string
	options   *ContainerOptions
}

// ContainerOptions are options for the Container resource
type ContainerOptions struct {
	Local      bool     `json:"local"`
	Exclusions []string `json:"exclusions"`
}

// NewContainer returns a configured Container resource for the scanner to scan
func NewContainer(location string, options *ContainerOptions) *Container {
	return &Container{
		location: location,
		options:  options,
	}
}

// Kind of resource (always returns Container here)
func (r *Container) Kind() string {
	return "Container"
}

// String representation of the resource
func (r *Container) String() string {
	return r.location
}

// Clone the resource to the desired local location and store the path
func (r *Container) Clone(path string) error {
	r.clonePath = path

	ctx := context.Background()

	sysCtx := &types.SystemContext{
		ArchitectureChoice: "amd64",
	}

	srcRef, err := docker.ParseReference("//" + r.location)
	if err != nil {
		logger.Error("Error parsing image reference: %v", err)
		return err
	}

	imageSource, err := srcRef.NewImageSource(ctx, sysCtx)
	if err != nil {
		logger.Error("Error creating image source: %v", err)
		return err
	}
	defer imageSource.Close()

	rawManifest, _, err := imageSource.GetManifest(ctx, nil)
	if err != nil {
		logger.Error("Error fetching manifest: %v", err)
		return err
	}

	imgManifest, err := manifest.Schema2FromManifest(rawManifest)
	if err != nil {
		logger.Error("Error parsing manifest: %v", err)
		return err
	}
	cache := blobinfocache.DefaultCache(sysCtx)
	layerDir := filepath.Join(path, "layers")
	err = os.MkdirAll(layerDir, 0755)
	if err != nil {
		logger.Error("Error creating layer directory: %v", err)
		return err
	}

	for _, layer := range imgManifest.LayersDescriptors {
		fmt.Printf("Downloading layer %s\n", layer.Digest)
		for _, skip := range r.options.Exclusions {
			if skip == layer.Digest.String() {
				println("Skipping layer")
				continue
			}
		}
		blobInfo := types.BlobInfo{
			Digest: layer.Digest,
			Size:   layer.Size,
		}
		layerBlob, _, err := imageSource.GetBlob(ctx, blobInfo, cache)
		if err != nil {
			logger.Error("Error downloading layer blob: %v", err)
			return err
		}
		defer layerBlob.Close()

		filePath := filepath.Join(layerDir, layer.Digest.Hex()+".tar.gz")
		file, err := os.Create(filePath)
		if err != nil {
			logger.Error("Error creating file: %v", err)
			return err
		}
		defer file.Close()

		_, err = io.Copy(file, layerBlob)
		if err != nil {
			logger.Error("Error writing layer to file: %v", err)
			return err
		}
		println(filePath)

	}
	return nil
}

// ClonePath returns where this repo has been cloned if cloned else ""
func (r *Container) ClonePath() string {
	return r.clonePath
}

// Depth returns the depth for things that have version control
func (r *Container) Depth() uint16 {
	return 0
}

// SetDepth allows you to adjust the depth for the resource
func (r *Container) SetDepth(depth uint16) {
	// no-op
}

// SetCloneTimeout lets you adjust the timeout before the clone aborts
func (r *Container) SetCloneTimeout(timeout time.Duration) {
	// no-op
}

// Since returns the date after which things should be scanned for things
// that have versions
func (r *Container) Since() string {
	return ""
}

// ReadFile provides a way to access values in the resource
func (r *Container) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(filepath.Join(r.ClonePath(), filepath.Clean(path)))
}

// Walk traverses the resource like a directory tree
func (r *Container) Walk(fn WalkFunc) error {
	// Handle if path is a file
	if fs.FileExists(r.ClonePath()) {
		file, err := os.Open(r.ClonePath())
		if err != nil {
			return err
		}
		defer file.Close()

		// path is empty because it's not in a directory
		return fn("", file)
	}

	return filepath.WalkDir(r.ClonePath(), func(path string, d iofs.DirEntry, err error) error {
		if err != nil {
			logger.Error("could not walk path: path=%q error=%q", path, err)
			return nil
		}

		relPath, err := filepath.Rel(r.ClonePath(), path)
		if err != nil {
			logger.Error("could generate relative path: path=%q error=%q", path, err)
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
			logger.Info("skipping symlink: path=%q", relPath)
			return nil
		}

		file, err := os.Open(filepath.Clean(path))
		if err != nil {
			logger.Error("could not open file: path=%q error=%q", relPath, err)
			return nil
		}
		defer file.Close()

		return fn(relPath, file)
	})
}
