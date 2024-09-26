package resource

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	iofs "io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/leaktk/scanner/pkg/fs"

	"github.com/leaktk/scanner/pkg/logger"
	"github.com/leaktk/scanner/pkg/response"

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
	manifest  *string
	labels    map[string]string
}

// ContainerImageOptions are options for the ContainerImage resource
type ContainerImageOptions struct {
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

// Contact Attempts to identify author information returing name and email if found
// The order was selected for most completeness with a preference to maintainer and OCI spec
// Returns the name and email
func (r *ContainerImage) Contact() (name string, email string) {
	if e, ok := r.labels["email"]; ok {
		email = strings.TrimSpace(e)
	}
	if authors, ok := r.labels["org.opencontainers.image.authors"]; ok {
		if match := extractRFC5322Mailbox(authors); match != nil {
			return match[0], match[1]
		}
		return strings.TrimSpace(authors), email
	}
	if maintainer, ok := r.labels["org.opencontainers.image.maintainers"]; ok {
		if match := extractRFC5322Mailbox(maintainer); match != nil {
			return match[0], match[1]
		}
		return strings.TrimSpace(maintainer), email
	}
	if maintainer, ok := r.labels["maintainer"]; ok {
		if match := extractRFC5322Mailbox(maintainer); match != nil {
			return match[0], match[1]
		}
		return strings.TrimSpace(maintainer), email
	}
	if author, ok := r.labels["author"]; ok {
		return strings.TrimSpace(author), email
	}
	return name, email
}

// Extracts RFC5322 style Mailboxes i.e "John Smith <jsmith@example.com>"
func extractRFC5322Mailbox(mailbox string) []string {
	for _, mb := range strings.Split(mailbox, ",") {
		re := regexp.MustCompile(`^(.*)\s<([^>]+)>$`)
		matches := re.FindStringSubmatch(mb)
		if len(matches) > 2 {
			return matches[1:]
		}
	}
	return nil
}

// Kind of resource (always returns ContainerImage here)
func (r *ContainerImage) Kind() string {
	return "ContainerImage"
}

// String representation of the resource
func (r *ContainerImage) String() string {
	return r.location
}

// Clone the resource to the desired clonePath location
func (r *ContainerImage) Clone(path string) error {
	err := os.MkdirAll(path, 0700)
	if err != nil {
		return fmt.Errorf("could not create clone directory: %v", err)
	}
	r.clonePath = path

	return r.cloneRemoteResource(path, r.location)
}

// cloneRemoteResource clones a remote resource ready for scanning.
func (r *ContainerImage) cloneRemoteResource(path string, resource string) error {

	ctx := context.Background()

	sysCtx := &types.SystemContext{}

	imgRef, err := docker.ParseReference("//" + resource)
	if err != nil {
		return fmt.Errorf("could not parse image reference: %v", err)
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
	if r.manifest == nil {
		// We only want the first manifest as it includes all of them
		err = r.writeFile("manifest.json", rawManifest)
		if err != nil {
			return fmt.Errorf("failed writing manifest to clonepath: %v", err)
		}
		stringManifest := string(rawManifest)
		r.manifest = &stringManifest
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
			logger.Info("manifest contains multiple options, defaulted to first (OS: %s, Arch: %s)",
				indexManifest.Manifests[index].Platform.OS, indexManifest.Manifests[index].Platform.Architecture)
		}
		imgRefString := imageSource.Reference().DockerReference().Name() + "@" + indexManifest.Manifests[index].Digest.String()

		return r.cloneRemoteResource(path, imgRefString)
	}

	img, err := imgRef.NewImage(ctx, sysCtx)
	if err != nil {
		log.Fatalf("could not load image to retrieve labels: %v", err)
	}
	defer img.Close()

	config, err := img.OCIConfig(ctx)
	if err != nil {
		log.Fatalf("could not get image config to retrieve labels: %v", err)
	}
	r.labels = config.Config.Labels

	configJSON, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to create string from configjson: %v", err)
	}
	err = r.writeFile("config.json", configJSON)
	if err != nil {
		return fmt.Errorf("failed to write config to clonepath: %v", err)
	}

	imgManifest, err := manifest.FromBlob(rawManifest, manifestType)
	if err != nil {
		return fmt.Errorf("could not parse manifest: %v", err)
	}

	cache := blobinfocache.DefaultCache(sysCtx)
	for _, layer := range imgManifest.LayerInfos() {

		if r.skipLayer(layer.Digest.Hex()) {
			continue
		}
		logger.Debug("downloading layer %s", layer.Digest.Hex())

		blobInfo := types.BlobInfo{
			Digest: layer.Digest,
			Size:   layer.Size,
		}
		layerBlob, _, err := imageSource.GetBlob(ctx, blobInfo, cache)
		if err != nil {
			return fmt.Errorf("could not download layer blob: %v", err)
		}

		err = r.extract(layerBlob, layer, path)
		if err != nil {
			return fmt.Errorf("could not decompress layer: %v", err)
		}
		err = layerBlob.Close()
		if err != nil {
			return fmt.Errorf("could not close layer: %v", err)
		}
	}

	return nil
}

func (r *ContainerImage) writeFile(filename string, content []byte) error {
	return os.WriteFile(filepath.Join(r.clonePath, filename), content, 0600)
}

// The decompression process is a little more involved so separated out.
func (r *ContainerImage) extract(t io.Reader, layer manifest.LayerInfo, path string) error {
	// The maximum file size should be less than 10x the layer size.
	size := layer.Size * 10
	layerRootDir := filepath.Join(r.ClonePath(), layer.Digest.Hex())
	layerDir := filepath.Join(path, layer.Digest.Hex())
	err := os.MkdirAll(layerDir, 0700)
	if err != nil {
		return fmt.Errorf("could not create layer directory: %v", err)
	}

	var tarReader *tar.Reader

	if strings.Contains(strings.ToLower(layer.MediaType), "gzip") {
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
		path, err := fs.CleanJoin(layerRootDir, header.Name)
		if err != nil {
			logger.Error("%v - skipped", err)
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
			logger.Error("%v", err)
			continue
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

// EnrichResult adds contextual information to each result
func (r *ContainerImage) EnrichResult(result *response.Result) *response.Result {
	if hash, file, found := strings.Cut(result.Location.Path, string(os.PathSeparator)); found {
		result.Location.Version = hash
		result.Location.Path = file
		result.Kind = response.ContainerLayerResultKind
	} else {
		result.Kind = response.ContainerMetdataResultKind
	}

	result.Notes = r.labels
	result.Contact.Name, result.Contact.Email = r.Contact()

	return result
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

// skipLayer checks if the digest is in the exclusion list and returns true if it is
func (r *ContainerImage) skipLayer(digest string) bool {
	for _, exclude := range r.options.Exclusions {
		if exclude == digest {
			logger.Info("skipping layer %s", digest)
			return true
		}
	}
	return false
}

// ReadFile provides a way to access values in the resource
func (r *ContainerImage) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(filepath.Join(r.ClonePath(), filepath.Clean(path)))
}

// Walk traverses the resource like a directory tree
func (r *ContainerImage) Walk(fn WalkFunc) error {
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
