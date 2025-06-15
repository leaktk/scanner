package gitleaks

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/zricethezav/gitleaks/v8/detect"
	"github.com/zricethezav/gitleaks/v8/report"
	"github.com/zricethezav/gitleaks/v8/sources"
)

var defaultRemote = &sources.RemoteInfo{}

// GitScanOpts configures ScanGit
type GitScanOpts struct {
	Branch   string
	Depth    int
	Remote   *sources.RemoteInfo
	Since    string
	Staged   bool
	Unstaged bool
}

// JSONScanOpts configures ScanJSON
type JSONScanOpts struct {
	FetchURLPatterns []string
}

// URLScanOpts configures ScanURL
type URLScanOpts struct {
	FetchURLPatterns []string
}

func ScanReader(ctx context.Context, detector *detect.Detector, reader io.Reader) ([]report.Finding, error) {
	return detector.DetectSource(
		ctx,
		&sources.File{
			Config:          &detector.Config,
			Content:         reader,
			MaxArchiveDepth: detector.MaxArchiveDepth,
		},
	)
}

func ScanURL(ctx context.Context, detector *detect.Detector, rawURL string, opts URLScanOpts) ([]report.Finding, error) {
	return detector.DetectSource(
		ctx,
		&URL{
			Config:           &detector.Config,
			FetchURLPatterns: opts.FetchURLPatterns,
			MaxArchiveDepth:  detector.MaxArchiveDepth,
			RawURL:           rawURL,
		},
	)
}

func ScanJSON(ctx context.Context, detector *detect.Detector, data string, opts JSONScanOpts) ([]report.Finding, error) {
	return detector.DetectSource(
		ctx,
		&JSON{
			Config:           &detector.Config,
			FetchURLPatterns: opts.FetchURLPatterns,
			MaxArchiveDepth:  detector.MaxArchiveDepth,
			RawMessage:       json.RawMessage(data),
		},
	)
}

func ScanFiles(ctx context.Context, detector *detect.Detector, path string) ([]report.Finding, error) {
	return detector.DetectSource(
		ctx,
		&sources.Files{
			Config:          &detector.Config,
			FollowSymlinks:  detector.FollowSymlinks,
			Path:            path,
			Sema:            detector.Sema,
			MaxArchiveDepth: detector.MaxArchiveDepth,
		},
	)
}

func ScanGit(ctx context.Context, detector *detect.Detector, repo string, opts GitScanOpts) ([]report.Finding, error) {
	var err error
	var gitCmd *sources.GitCmd
	var remote *sources.RemoteInfo

	if opts.Unstaged || opts.Staged {
		if gitCmd, err = sources.NewGitDiffCmd(repo, opts.Staged); err != nil {
			return nil, fmt.Errorf("could not create git diff cmd: %w", err)
		}
	} else {
		logOpts := []string{"--full-history", "--ignore-missing"}

		if len(opts.Since) > 0 {
			logOpts = append(logOpts, "--since")
			logOpts = append(logOpts, opts.Since)
		}

		if opts.Depth > 0 {
			logOpts = append(logOpts, "--max-count")
			logOpts = append(logOpts, fmt.Sprint(opts.Depth))
		}

		if len(opts.Branch) > 0 {
			logOpts = append(logOpts, opts.Branch)
		} else {
			logOpts = append(logOpts, "--all")
		}

		if shallowCommits := shallowCommits(repo); len(shallowCommits) > 0 {
			logOpts = append(logOpts, "--not")
			logOpts = append(logOpts, shallowCommits...)
		}

		if gitCmd, err = sources.NewGitLogCmd(repo, strings.Join(logOpts, " ")); err != nil {
			return nil, fmt.Errorf("could not create git log cmd: %w", err)
		}
	}

	if opts.Remote != nil {
		remote = opts.Remote
	} else {
		remote = defaultRemote
	}

	return detector.DetectSource(
		ctx,
		&sources.Git{
			Cmd:             gitCmd,
			Config:          &detector.Config,
			Remote:          remote,
			Sema:            detector.Sema,
			MaxArchiveDepth: detector.MaxArchiveDepth,
		},
	)
}

func shallowCommits(repo string) []string {
	var shallowCommits []string

	// TODO: replace with git command to get repo-dir
	data, err := os.ReadFile(filepath.Join(repo, "shallow"))
	if err != nil {
		return shallowCommits
	}

	for _, shallowCommit := range strings.Split(string(data), "\n") {
		if len(shallowCommit) > 0 {
			shallowCommits = append(shallowCommits, shallowCommit)
		}
	}

	return shallowCommits
}
