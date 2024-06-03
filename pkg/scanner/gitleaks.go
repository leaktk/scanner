package scanner

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/zricethezav/gitleaks/v8/detect"
	"github.com/zricethezav/gitleaks/v8/sources"

	"github.com/leaktk/scanner/pkg/fs"
	"github.com/leaktk/scanner/pkg/logger"
	"github.com/leaktk/scanner/pkg/resource"
)

// Gitleaks wraps gitleaks as a scanner backend
type Gitleaks struct {
	patterns *Patterns
}

// NewGitLeaks returns a configured gitleaks backend instance
func NewGitleaks(patterns *Patterns) *Gitleaks {
	return &Gitleaks{
		patterns: patterns,
	}
}

// Name returns the human readable name of the backend for logging details
func (g *Gitleaks) Name() string {
	return "Gitleaks"
}

// newDetector creates and configures a detector object for this resource
func (g *Gitleaks) newDetector(scanResource resource.Resource) (*detect.Detector, error) {
	cfg, err := g.patterns.Gitleaks()

	if err != nil {
		return nil, err
	}

	detector := detect.NewDetector(*cfg)
	detector.FollowSymlinks = false
	detector.IgnoreGitleaksAllow = false
	detector.MaxTargetMegaBytes = 0
	detector.NoColor = true
	detector.Redact = 0
	detector.Verbose = false

	gitleaksIgnorePath := filepath.Join(scanResource.ClonePath(), ".gitleaksignore")
	if fs.FileExists(gitleaksIgnorePath) {
		if err = detector.AddGitleaksIgnore(gitleaksIgnorePath); err != nil {
			return nil, fmt.Errorf("could not call AddGitleaksIgnore (%v)", err)
		}
	}

	gitleaksBaselinePath := filepath.Join(scanResource.ClonePath(), ".gitleaksbaseline")
	if fs.FileExists(gitleaksBaselinePath) {
		if err = detector.AddBaseline(gitleaksBaselinePath, scanResource.ClonePath()); err != nil {
			return nil, fmt.Errorf("could not call AddBaseline (%v)", err)
		}
	}

	rawClonedConfig, err := scanResource.ReadFile(".gitleaks.toml")
	if err == nil {
		clonedConfig, err := ParseGitleaksConfig(string(rawClonedConfig))

		if err != nil {
			detector.Config.Allowlist.Commits = append(detector.Config.Allowlist.Commits, clonedConfig.Allowlist.Commits...)
			detector.Config.Allowlist.Paths = append(detector.Config.Allowlist.Paths, clonedConfig.Allowlist.Paths...)
			detector.Config.Allowlist.Regexes = append(detector.Config.Allowlist.Regexes, clonedConfig.Allowlist.Regexes...)
		} else {
			logger.Error("could not load cloned .gitleaks.toml: resource_id=%q error=%q", scanResource.ID(), err)
		}
	}

	return detector, nil
}

func (g *Gitleaks) shallowCommits(scanResource resource.Resource) []string {
	if gitRepo, ok := scanResource.(*resource.GitRepo); ok {
		shallowFilePath := filepath.Join(gitRepo.GitDirPath(), "shallow")
		data, err := os.ReadFile(filepath.Clean(shallowFilePath))

		if err == nil {
			return strings.Split(string(data), "\n")
		}
	}

	return []string{}
}

// Scan does the gitleaks scan on the resource
func (g *Gitleaks) Scan(scanResource resource.Resource) ([]*Result, error) {
	gitLogOpts := []string{"--full-history", "--all"}

	if len(scanResource.Since()) > 0 {
		gitLogOpts = append(gitLogOpts, "--since")
		gitLogOpts = append(gitLogOpts, scanResource.Since())
	}

	// Should be the last set of args
	if shallowCommits := g.shallowCommits(scanResource); len(shallowCommits) > 0 {
		gitLogOpts = append(gitLogOpts, "--not")
		gitLogOpts = append(gitLogOpts, shallowCommits...)
	}

	gitCmd, err := sources.NewGitLogCmd(scanResource.ClonePath(), strings.Join(gitLogOpts, " "))

	if err != nil {
		return nil, err
	}

	detector, err := g.newDetector(scanResource)
	if err != nil {
		return nil, err
	}

	findings, err := detector.DetectGit(gitCmd)
	results := make([]*Result, len(findings))

	for i, finding := range findings {
		results[i] = &Result{
			// Be careful changing how this is generated, this could result in
			// duplicate alerts
			ID: ResultID(
				// What: Uniquely identify the kind of thing that's being scanned
				GitCommitResultKind,
				scanResource.String(),

				// Where: Uniquely identify where in that resource it was being scanned
				finding.Commit,
				finding.File,
				fmt.Sprint(finding.StartLine),
				fmt.Sprint(finding.StartColumn),
				fmt.Sprint(finding.EndLine),
				fmt.Sprint(finding.EndColumn),

				// How: Uniquely identify what was used to find it
				finding.RuleID,
			),
			Kind:    GitCommitResultKind,
			Secret:  finding.Secret,
			Match:   finding.Match,
			Entropy: finding.Entropy,
			Date:    finding.Date,
			Notes: map[string]string{
				"message": finding.Message,
			},
			Contact: Contact{
				Name:  finding.Author,
				Email: finding.Email,
			},
			Rule: Rule{
				ID:          finding.RuleID,
				Description: finding.Description,
				Tags:        finding.Tags,
			},
			Location: Location{
				Version: finding.Commit,
				Path:    finding.File,
				Start: Point{
					Line:   finding.StartLine,
					Column: finding.StartColumn,
				},
				End: Point{
					Line:   finding.EndLine,
					Column: finding.EndColumn,
				},
			},
		}
	}

	return results, err
}
