package resource

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"github.com/leaktk/scanner/pkg/fs"
	"github.com/leaktk/scanner/pkg/logger"
	"io"
	iofs "io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/containers/image/v5/docker"
	"github.com/containers/image/v5/manifest"
	"github.com/containers/image/v5/pkg/blobinfocache"
	"github.com/containers/image/v5/types"
)

// ContainerImage provides a pull and hold a container
type ContainerImage struct {
	// Provide common helper functions
	BaseResource
	clonePath string
	location  string
	options   *ContainerImageOptions
}

// ContainerImageOptions are options for the ContainerImage resource
type ContainerImageOptions struct {
	Local      bool     `json:"local"`
	Exclusions []string `json:"exclusions"`
	Arch       string   `json:"arch"`
}

// NewContainerImage returns a configured ContainerImage resource for the scanner to scan
func NewContainerImage(location string, options *ContainerImageOptions) *ContainerImage {
	return &ContainerImage{
		location: location,
		options:  options,
	}
}

// Kind of resource (always returns ContainerImage here)
func (r *ContainerImage) Kind() string {
	return "ContainerImage"
}

// String representation of the resource
func (r *ContainerImage) String() string {
	return r.location
}

// Clone the resource to the desired local location and store the path
func (r *ContainerImage) Clone(path string) error {
	r.clonePath = path
	if r.options.Local {
		return r.cloneLocalResource(path, r.location)
	}
	return r.cloneRemoteResource(path, r.location)
}

func (r *ContainerImage) cloneLocalResource(clonePath string, location string) error {
	// Do local stuff here - likely just decompress/untar?
	return nil
}

func (r *ContainerImage) cloneRemoteResource(path string, resource string) error {

	ctx := context.Background()

	sysCtx := &types.SystemContext{}

	imgRef, err := docker.ParseReference("//" + resource)
	if err != nil {
		return fmt.Errorf("Error parsing image reference: %v", err)
	}

	imageSource, err := imgRef.NewImageSource(ctx, sysCtx)
	if err != nil {
		return fmt.Errorf("could not create image source: %v", err)
	}
	defer imageSource.Close()

	rawManifest, manifestType, err := imageSource.GetManifest(ctx, nil)
	if err != nil {
		return fmt.Errorf("could not fetch manifest: %v", err)
	}

	if manifestType == manifest.DockerV2ListMediaType { // multiple entries select first
		var indexManifest manifest.Schema2List
		index := 0
		if err := json.Unmarshal(rawManifest, &indexManifest); err != nil {
			return fmt.Errorf("could not unmarshal manifest: %v", err)
		}
		if r.options.Arch != "" {
			for i, m := range indexManifest.Manifests {
				if m.Platform.Architecture == r.options.Arch {
					index = i
					logger.Info("selected first %s container", r.options.Arch)
					break
				}
			}
		} else {
			logger.Info("manifest contains multiple options, defaulted to first")
		}
		imgRef := imageSource.Reference().DockerReference().Name() + "@" + indexManifest.Manifests[index].Digest.String()

		return r.cloneRemoteResource(path, imgRef)

	}

	imgManifest, err := manifest.Schema2FromManifest(rawManifest)
	if err != nil {
		return fmt.Errorf("could not parse manifest: %v", err)
	}

	cache := blobinfocache.DefaultCache(sysCtx)
	for _, layer := range imgManifest.LayersDescriptors {
		for _, skip := range r.options.Exclusions {
			if skip == layer.Digest.String() {
				logger.Debug("skipping layer %s", layer.Digest.String())
				continue
			}
		}
		logger.Debug("downloading layer %s", layer.Digest.String())

		layerDir := filepath.Join(path, layer.Digest.Hex())
		err = os.MkdirAll(layerDir, 0700)
		if err != nil {
			return fmt.Errorf("could not create layer directory: %v", err)
		}

		blobInfo := types.BlobInfo{
			Digest: layer.Digest,
			Size:   layer.Size,
		}
		layerBlob, _, err := imageSource.GetBlob(ctx, blobInfo, cache)
		if err != nil {
			return fmt.Errorf("could not download layer blob: %v", err)
		}
		defer layerBlob.Close()

		err = r.decompress(layerBlob, layer.Digest.Hex(), layer.MediaType, layer.Size)
		if err != nil {
			return fmt.Errorf("could not decompress layer: %v", err)
		}
	}
	return nil

}

// The decompression process is a little more involved so separated out.
func (r *ContainerImage) decompress(t io.Reader, layer string, mediaType string, size int64) error {
	// The maximum file size should be less than 10x the layer size.
	size = size * 10
	path := filepath.Join(r.ClonePath(), layer)

	var tarReader *tar.Reader

	if strings.Contains(strings.ToLower(mediaType), "gzip") {
		gzReader, err := gzip.NewReader(t)
		if err != nil {
			return err
		}
		tarReader = tar.NewReader(gzReader)
		defer gzReader.Close()
	} else {
		tarReader = tar.NewReader(t)
	}

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		path, err := sanitizePath(path, header.Name)
		if err != nil {
			logger.Error("%s - skipped", err)
			continue
		}
		info := header.FileInfo()
		if info.IsDir() {
			if err = os.MkdirAll(path, 0700); err != nil {
				return err
			}
			continue
		}

		file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600) // #nosec G304
		if err != nil {
			return err
		}
		defer file.Close()

		// Copy a maximum number of bytes (layer size * 10) so we do not get "bombs". It is unlikely that a file
		// with significant entropy will be compressed more than 10x. We can review this.
		n, err := io.CopyN(file, tarReader, size)
		if err != nil && err != io.EOF {
			return fmt.Errorf("could not copy file to disk: %v", err)
		}
		if n >= size {
			logger.Warning("copying file %s did not finish due to max file size: %v", file.Name(), err)
		}
	}
	return nil
}

// ClonePath returns where this repo has been cloned if cloned else ""
func (r *ContainerImage) ClonePath() string {
	return r.clonePath
}

// Depth returns the depth for things that have version control
func (r *ContainerImage) Depth() uint16 {
	return 0
}

// SetDepth allows you to adjust the depth for the resource
func (r *ContainerImage) SetDepth(depth uint16) {
	// no-op
}

// SetCloneTimeout lets you adjust the timeout before the clone aborts
func (r *ContainerImage) SetCloneTimeout(timeout time.Duration) {
	// no-op
}

// Since returns the date after which things should be scanned for things
// that have versions
func (r *ContainerImage) Since() string {
	return ""
}

// ReadFile provides a way to access values in the resource
func (r *ContainerImage) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(filepath.Join(r.ClonePath(), filepath.Clean(path)))
}

// Walk traverses the resource like a directory tree
func (r *ContainerImage) Walk(fn WalkFunc) error {
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

func sanitizePath(destination string, filePath string) (string, error) {
	destPath := filepath.Join(destination, filePath)
	if !strings.HasPrefix(destPath, filepath.Clean(destination)) {
		return "", fmt.Errorf("illegal file path: %s", filePath)
	}
	return destPath, nil
}
