package scanner

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/zricethezav/gitleaks/v8/detect"
	"github.com/zricethezav/gitleaks/v8/sources"

	"github.com/leaktk/scanner/pkg/resource"
)

// TODO: include the repo's .gitleaks.toml's allowlist and figure out
// if we wan't to support any extra rules for when the time comes to use
// this as a pre-commit hook. We could just strip the tags on them or and
// limit how many rules we allow or something similar.

func fileExists(fileName string) bool {
	info, err := os.Stat(fileName)
	if err != nil && !os.IsNotExist(err) {
		return false
	}

	if info != nil && err == nil {
		if !info.IsDir() {
			return true
		}
	}
	return false
}

// Gitleaks wraps gitleaks as a scanner backend
type Gitleaks struct {
	patterns   *Patterns
	gitLogOpts string
}

func NewGitleaks(patterns *Patterns) *Gitleaks {
	return &Gitleaks{
		patterns:   patterns,
		gitLogOpts: "", // Custom git log flags we can pass
	}
}

// newDetector creates and configures a detector object for this resource
func (g *Gitleaks) newDetector(scanResource resource.Resource) (*detect.Detector, error) {
	gitleaksConfig, err := g.patterns.Gitleaks()

	if err != nil {
		return nil, err
	}

	detector := detect.NewDetector(*gitleaksConfig)
	detector.FollowSymlinks = false
	detector.IgnoreGitleaksAllow = false
	detector.MaxTargetMegaBytes = 0
	detector.NoColor = true
	detector.Redact = 0
	detector.Verbose = false

	gitleaksIgnorePath := filepath.Join(scanResource.ClonePath(), ".gitleaksignore")
	if fileExists(gitleaksIgnorePath) {
		if err = detector.AddGitleaksIgnore(gitleaksIgnorePath); err != nil {
			return nil, fmt.Errorf("could not call AddGitleaksIgnore (%v)", err)
		}
	}

	gitleaksBaselinePath := filepath.Join(scanResource.ClonePath(), ".gitleaksbaseline")
	if fileExists(gitleaksBaselinePath) {
		if err = detector.AddBaseline(gitleaksBaselinePath, scanResource.ClonePath()); err != nil {
			return nil, fmt.Errorf("could not call AddBaseline (%v)", err)
		}
	}

	return detector, nil
}

// Scan does the gitleaks scan on the resource
func (g *Gitleaks) Scan(scanResource resource.Resource) ([]*Result, error) {
  // TODO: make sure to add log opts to ignore any grafed commits when the scanner has done the clone
	gitCmd, err := sources.NewGitLogCmd(scanResource.ClonePath(), g.gitLogOpts)

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
      // TODO: We want to strike a balance between stable fields and unique
      // ones. We want to avoid re-reporting when there's been a small change.
			ID: ResultID(
				scanResource.Kind(),
				finding.RuleID,
				scanResource.String(),
				finding.Commit,
				finding.File,
				fmt.Sprint(finding.StartLine),
				fmt.Sprint(finding.StartColumn),
				fmt.Sprint(finding.EndLine),
				fmt.Sprint(finding.EndColumn),
			),
			Kind:    scanResource.Kind(),
			Secret:  finding.Secret,
			Match:   finding.Match,
			Entropy: finding.Entropy,
			Date:    finding.Date,
			Notes:   fmt.Sprintf("commit message: %s", finding.Message),
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
