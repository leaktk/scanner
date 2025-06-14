package resource

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/leaktk/leaktk/pkg/fs"
	"github.com/leaktk/leaktk/pkg/logger"
	"github.com/leaktk/leaktk/pkg/proto"
	"github.com/leaktk/leaktk/version"

	"github.com/containers/image/v5/docker"
	"github.com/containers/image/v5/manifest"
	"github.com/containers/image/v5/pkg/blobinfocache"
	"github.com/containers/image/v5/types"
	"github.com/klauspost/compress/zstd"
)

var rfc5322Regexp = regexp.MustCompile(`^(.*)\s<([^>]+)>$`)

// Extracts RFC5322 style Mailboxes i.e "John Smith <jsmith@example.com>"
func extractRFC5322Mailbox(mailbox string) []string {
	for _, mb := range strings.Split(mailbox, ",") {
		matches := rfc5322Regexp.FindStringSubmatch(mb)
		if len(matches) > 2 {
			return matches[1:]
		}
	}
	return nil
}

// ContainerImageOptions are options for the ContainerImage resource
type ContainerImageOptions struct {
	// A preferred arch, if it exists - defaults to first
	Arch string `json:"arch"`
	// Set the number of layers to download, counting from the top down.
	Depth int `json:"depth"`
	// A list of layer hashes to exclude from clone and scan
	Exclusions []string `json:"exclusions"`
	// The scan priority
	Priority int `json:"priority"`
	// Only scan since this date
	Since string `json:"since"`
}

// Contact Attempts to identify author information returning name and email if found
// The order was selected for most completeness with a preference to maintainer and OCI spec
// Returns the name and email
func (r *ContainerImage) Contact() proto.Contact {
	var email string
	if e, ok := r.labels["email"]; ok {
		email = strings.TrimSpace(e)
	}
	if authors, ok := r.labels["org.opencontainers.image.authors"]; ok {
		if match := extractRFC5322Mailbox(authors); match != nil {
			return proto.Contact{Name: match[0], Email: match[1]}
		}
		return proto.Contact{Name: strings.TrimSpace(authors), Email: email}
	}
	if maintainer, ok := r.labels["org.opencontainers.image.maintainers"]; ok {
		if match := extractRFC5322Mailbox(maintainer); match != nil {
			return proto.Contact{Name: match[0], Email: match[1]}
		}
		return proto.Contact{Name: strings.TrimSpace(maintainer), Email: email}
	}
	if maintainer, ok := r.labels["maintainer"]; ok {
		if match := extractRFC5322Mailbox(maintainer); match != nil {
			return proto.Contact{Name: match[0], Email: match[1]}
		}
		return proto.Contact{Name: strings.TrimSpace(maintainer), Email: email}
	}
	if author, ok := r.labels["author"]; ok {
		return proto.Contact{Name: strings.TrimSpace(author), Email: email}
	}
	return proto.Contact{Email: email}
}

// TODO: Just scan these on the fly using readers
func (r *ContainerImage) Clone(path string) error {
	err := os.MkdirAll(path, 0700)
	if err != nil {
		return fmt.Errorf("could not create clone directory: %v", err)
	}

	r.path = path
	if r.cloneTimeout > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), r.cloneTimeout)
		defer cancel()
		return r.cloneRemoteResource(ctx, path, r.location)
	}

	ctx := context.Background()
	return r.cloneRemoteResource(ctx, path, r.location)
}

// cloneRemoteResource clones a remote resource ready for scanning.
func (r *ContainerImage) cloneRemoteResource(ctx context.Context, path string, resource string) error {
	sysCtx := &types.SystemContext{
		DockerRegistryUserAgent: version.GlobalUserAgent,
	}

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
					r.Info(logger.CloneDetail, "selected first %s container", r.options.Arch)
					break
				}
			}
		} else {
			r.Info(logger.CloneDetail, "manifest contains multiple options, defaulted to first (OS: %s, Arch: %s)",
				indexManifest.Manifests[index].Platform.OS, indexManifest.Manifests[index].Platform.Architecture)
		}
		imgRefString := imageSource.Reference().DockerReference().Name() + "@" + indexManifest.Manifests[index].Digest.String()

		return r.cloneRemoteResource(ctx, path, imgRefString)
	}

	img, err := imgRef.NewImage(ctx, sysCtx)
	if err != nil {
		return fmt.Errorf("could not load image to retrieve labels: %v", err)
	}
	defer img.Close()

	config, err := img.OCIConfig(ctx)
	if err != nil {
		return fmt.Errorf("could not get image config to retrieve labels: %v", err)
	}

	var layerHistoryDates []*time.Time
	for _, layerHistory := range config.History {
		if !layerHistory.EmptyLayer {
			layerHistoryDates = append(layerHistoryDates, layerHistory.Created)
		}
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
	layers := imgManifest.LayerInfos()
	since := r.sinceTime()
	layers, layerHistoryDates = r.layerDepth(layers, layerHistoryDates)
	for i, layer := range layers {
		if since != nil && layerHistoryDates != nil {
			if layerHistoryDates[i].Before(*since) {
				r.Info(logger.CloneDetail, "layer older than provided date, skipping layer %s", layer.Digest.Hex())
				continue
			}
		}
		if r.skipLayer(layer.Digest.Hex()) {
			continue
		}
		r.Debug(logger.CloneDetail, "downloading layer %s", layer.Digest.Hex())

		blobInfo := types.BlobInfo{
			Digest: layer.Digest,
			Size:   layer.Size,
		}
		layerBlob, _, err := imageSource.GetBlob(ctx, blobInfo, cache)
		if err != nil {
			return fmt.Errorf("could not download layer blob: %v", err)
		}

		err = r.extractLayer(layerBlob, layer, path)
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
	return os.WriteFile(filepath.Join(r.path, filename), content, 0600)
}

func (r *ContainerImage) copyN(dst string, src io.Reader, n int64) error {
	file, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600) // #nosec G304
	if err != nil {
		return err
	}
	defer file.Close()

	// Copy a maximum number of bytes (layer size * 10) so we do not get "bombs". It is unlikely that a file
	// with significant entropy will be compressed more than 10x. We can review this.
	written, err := io.CopyN(file, src, n)
	if err != nil && err != io.EOF {
		return err
	}
	if written >= n {
		r.Warning(logger.CloneDetail, "copying file %s did not finish due to max file size: %v", file.Name(), err)
	}
	return nil
}

// The decompression process is a little more involved so separated out.
func (r *ContainerImage) extractLayer(t io.Reader, layer manifest.LayerInfo, path string) error {
	// The maximum file size should be less than 10x the layer size.
	size := layer.Size * 10
	layerRootDir := filepath.Join(r.Path(), layer.Digest.Hex())
	layerDir := filepath.Join(path, layer.Digest.Hex())
	err := os.MkdirAll(layerDir, 0700)
	if err != nil {
		return fmt.Errorf("could not create layer directory: %v", err)
	}

	var tarReader *tar.Reader

	if strings.HasSuffix(strings.ToLower(layer.MediaType), "gzip") {
		gzReader, err := gzip.NewReader(t)
		if err != nil {
			return fmt.Errorf("could not create gzip reader: %v", err)
		}
		tarReader = tar.NewReader(gzReader)
		defer gzReader.Close()
	} else if strings.HasSuffix(strings.ToLower(layer.MediaType), "zstd") {
		zstdReader, err := zstd.NewReader(t)
		if err != nil {
			return fmt.Errorf("could not create zstd reader: %v", err)
		}
		tarReader = tar.NewReader(zstdReader)
		defer zstdReader.Close()
	} else {
		tarReader = tar.NewReader(t)
	}

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return fmt.Errorf("could not extract tar: %v", err)
		}
		path, err := fs.CleanJoin(layerRootDir, header.Name)
		if err != nil {
			r.Error(logger.CloneError, "%v - skipped", err)
			continue
		}
		info := header.FileInfo()
		if info.IsDir() {
			if err = os.MkdirAll(path, 0700); err != nil {
				return fmt.Errorf("could not create directory: %v", err)
			}
			continue
		}
		if info.Mode()&os.ModeSymlink != 0 {
			r.Warning(logger.CloneDetail, "skipping file that is a symlink: %s", info.Name())
			continue
		}

		err = r.copyN(path, tarReader, size)
		if err != nil {
			r.Error(logger.CloneError, "could not create/write file: %v", err)
			// Try others, in case its unsupported file name for the filesystem etc.
			continue
		}
	}
	return nil
}

// layerDepth returns the layers to scan based on the depth provided
func (r *ContainerImage) layerDepth(layers []manifest.LayerInfo, dates []*time.Time) ([]manifest.LayerInfo, []*time.Time) {
	if len(layers) != len(dates) {
		// if our history length is different to our layer length, drop it.
		dates = nil
	}

	if r.Depth() == 0 {
		return layers, dates
	}

	sliceStart := len(layers) - int(r.Depth())
	if sliceStart < 0 {
		return layers, dates
	}

	return layers[sliceStart:], dates[sliceStart:]
}

// EnrichResult adds contextual information to each result
func (r *ContainerImage) EnrichResult(result *proto.Result) *proto.Result {
	if hash, file, found := strings.Cut(result.Location.Path, string(os.PathSeparator)); found {
		result.Location.Version = hash
		result.Location.Path = file
		result.Kind = proto.ContainerLayerResultKind
	} else {
		result.Kind = proto.ContainerMetdataResultKind
	}

	result.Notes = r.labels
	result.Contact = r.Contact()

	return result
}

func (r *ContainerImage) sinceTime() *time.Time {
	if len(r.options.Since) > 0 {
		date, err := time.Parse("2006-01-02", r.options.Since)
		if err != nil {
			r.Error(logger.CloneError, "could not parse since time: %v", err)
			return nil
		}

		return &date
	}

	return nil
}

// skipLayer checks if the digest is in the exclusion list and returns true if it is
func (r *ContainerImage) skipLayer(digest string) bool {
	for _, exclude := range r.options.Exclusions {
		if exclude == digest {
			r.Info(logger.CloneDetail, "layer in exclusion list, skipping layer %s", digest)
			return true
		}
	}
	return false
}
