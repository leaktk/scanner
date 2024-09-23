package scanner

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/h2non/filetype"
	"github.com/zricethezav/gitleaks/v8/detect"
	"github.com/zricethezav/gitleaks/v8/report"
	"github.com/zricethezav/gitleaks/v8/sources"

	"github.com/leaktk/scanner/pkg/fs"
	"github.com/leaktk/scanner/pkg/id"
	"github.com/leaktk/scanner/pkg/logger"
	"github.com/leaktk/scanner/pkg/resource"
)

const (
	chunkSize = 1024 * 1024 // 1 MiB
)

// Gitleaks wraps gitleaks as a scanner backend
type Gitleaks struct {
	patterns *Patterns
}

// NewGitleaks returns a configured gitleaks backend instance
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

	// TODO: move this to scanResource.ReadFile and have JSONData.Clone not write files to disk
	gitleaksIgnorePath := filepath.Join(scanResource.ClonePath(), ".gitleaksignore")
	if fs.FileExists(gitleaksIgnorePath) {
		if err = detector.AddGitleaksIgnore(gitleaksIgnorePath); err != nil {
			return nil, fmt.Errorf("could not add gitleaks ignore: error=%q", err)
		}
	}

	// TODO: move this to scanResource.ReadFile and have JSONData.Clone not write files to disk
	gitleaksBaselinePath := filepath.Join(scanResource.ClonePath(), ".gitleaksbaseline")
	if fs.FileExists(gitleaksBaselinePath) {
		if err = detector.AddBaseline(gitleaksBaselinePath, scanResource.ClonePath()); err != nil {
			return nil, fmt.Errorf("could not add baseline: error=%q", err)
		}
	}

	rawClonedConfig, err := scanResource.ReadFile(".gitleaks.toml")
	if err == nil {
		clonedConfig, err := ParseGitleaksConfig(string(rawClonedConfig))

		if err != nil {
			logger.Error("could not load cloned .gitleaks.toml: resource_id=%q error=%q", scanResource.ID(), err)
		} else {
			logger.Debug("loading cloned .gitleaks.toml")
			detector.Config.Allowlist.Commits = append(detector.Config.Allowlist.Commits, clonedConfig.Allowlist.Commits...)
			detector.Config.Allowlist.Paths = append(detector.Config.Allowlist.Paths, clonedConfig.Allowlist.Paths...)
			detector.Config.Allowlist.Regexes = append(detector.Config.Allowlist.Regexes, clonedConfig.Allowlist.Regexes...)
		}
	} else {
		logger.Debug("no cloned .gitleaks.toml")
	}

	return detector, nil
}

// gitScan handles when the resource is a gitRepo type
func (g *Gitleaks) gitScan(detector *detect.Detector, gitRepo *resource.GitRepo) ([]report.Finding, error) {
	gitLogOpts := []string{"--full-history", "--all", "--ignore-missing"}

	if len(gitRepo.Since()) > 0 {
		gitLogOpts = append(gitLogOpts, "--since")
		gitLogOpts = append(gitLogOpts, gitRepo.Since())
	}

	// Should be the last set of args
	if shallowCommits := gitRepo.ShallowCommits(); len(shallowCommits) > 0 {
		gitLogOpts = append(gitLogOpts, "--not")
		gitLogOpts = append(gitLogOpts, shallowCommits...)
	}

	gitCmd, err := sources.NewGitLogCmd(gitRepo.ClonePath(), strings.Join(gitLogOpts, " "))
	if err != nil {
		return nil, err
	}

	return detector.DetectGit(gitCmd)
}

// walkScan is the default way to scan most resources
func (g *Gitleaks) walkScan(detector *detect.Detector, scanResource resource.Resource) ([]report.Finding, error) {
	var findings []report.Finding

	err := scanResource.Walk(func(path string, reader io.Reader) error {
		// Source: https://github.com/gitleaks/gitleaks/blob/master/detect/directory.go
		buf := make([]byte, chunkSize)
		totalLines := 0

		for {
			n, err := reader.Read(buf)
			if err != nil && err != io.EOF {
				logger.Error("could not read file: path=%q", path)
				return nil
			}

			if n == 0 {
				break
			}

			// TODO: optimization could be introduced here
			mimetype, err := filetype.Match(buf[:n])
			if err != nil {
				logger.Error("could not determine file type: path=%q", path)
				return nil
			}
			if mimetype.MIME.Type == "application" {
				logger.Error("skipping binary file: path=%q", path)
				return nil // skip binary files
			}

			// Count the number of newlines in this chunk
			linesInChunk := strings.Count(string(buf[:n]), "\n")
			totalLines += linesInChunk
			fragment := detect.Fragment{
				Raw:      string(buf[:n]),
				FilePath: path,
			}

			for _, finding := range detector.Detect(fragment) {
				// need to add 1 since line counting starts at 1
				finding.StartLine += (totalLines - linesInChunk) + 1
				finding.EndLine += (totalLines - linesInChunk) + 1
				findings = append(findings, finding)
			}
		}

		return nil
	})

	return findings, err
}

// Scan does the gitleaks scan on the resource
func (g *Gitleaks) Scan(scanResource resource.Resource) ([]*Result, error) {
	var findings []report.Finding
	var err error
	var resultKind string

	detector, err := g.newDetector(scanResource)
	if err != nil {
		return nil, err
	}

	switch scanResource := scanResource.(type) {
	case *resource.GitRepo:
		findings, err = g.gitScan(detector, scanResource)
		resultKind = GitCommitResultKind
	case *resource.JSONData:
		findings, err = g.walkScan(detector, scanResource)
		resultKind = JSONDataResultKind
	default:
		findings, err = g.walkScan(detector, scanResource)
		resultKind = GeneralResultKind
	}

	if err != nil {
		logger.Error("gitleaks error: error=%q", err)
	}

	results := make([]*Result, len(findings))

	for i, finding := range findings {
		notes := map[string]string{}

		switch scanResource.(type) {
		case *resource.GitRepo:
			notes["message"] = finding.Message
		}

		results[i] = &Result{
			// Be careful changing how this is generated, this could result in
			// duplicate alerts
			ID: id.ID(
				// What: Uniquely identify the kind of thing that's being scanned
				resultKind,
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
			Kind:    resultKind,
			Secret:  finding.Secret,
			Match:   finding.Match,
			Entropy: finding.Entropy,
			Date:    finding.Date,
			Notes:   notes,
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
