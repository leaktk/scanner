package gitleaks

import (
	"fmt"

	"github.com/zricethezav/gitleaks/v8/config"
	"github.com/zricethezav/gitleaks/v8/detect"

	"github.com/leaktk/leaktk/pkg/fs"
	"github.com/leaktk/leaktk/pkg/logger"
)

type DetectorOpts struct {
	AdditionalConfig string
	BaselinePath     string
	IgnorePath       string
	MaxArchiveDepth  int
	MaxDecodeDepth   int
	SourcePath       string
}

func NewDetector(cfg config.Config, opts DetectorOpts) (*detect.Detector, error) {
	detector := detect.NewDetector(cfg)
	detector.FollowSymlinks = false
	detector.IgnoreGitleaksAllow = false
	detector.MaxArchiveDepth = opts.MaxArchiveDepth
	detector.MaxDecodeDepth = opts.MaxDecodeDepth
	detector.MaxTargetMegaBytes = 0
	detector.NoColor = true
	detector.Redact = 0
	detector.Verbose = false

	if fs.FileExists(opts.IgnorePath) {
		if err := detector.AddGitleaksIgnore(opts.IgnorePath); err != nil {
			return nil, fmt.Errorf("could not add gitleaksignore: %w", err)
		}
	}

	if fs.FileExists(opts.BaselinePath) {
		if err := detector.AddBaseline(opts.BaselinePath, opts.SourcePath); err != nil {
			return nil, fmt.Errorf("could not add baseline: %w", err)
		}
	}

	if len(opts.AdditionalConfig) > 0 {
		additionalConfig, err := ParseConfig(opts.AdditionalConfig)

		if err != nil {
			return nil, fmt.Errorf("could not parse additional config: %w", err)
		} else {
			logger.Debug("loading additional config")
			detector.Config.Allowlists = append(detector.Config.Allowlists, additionalConfig.Allowlists...)
		}
	} else {
		logger.Debug("no additional config")
	}

	return detector, nil
}
