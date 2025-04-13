package resource

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/leaktk/scanner/pkg/response"

	"github.com/leaktk/scanner/pkg/logger"
)

// Configure git env
func init() {
	// Make sure git never prompts for a password
	if err := os.Setenv("GIT_TERMINAL_PROMPT", "0"); err != nil {
		logger.Error("could not set GIT_TERMINAL_PROMPT=0: %w", err)
	}

	// Make sure replace is disabled so we can scan all of the refs
	if err := os.Setenv("GIT_NO_REPLACE_OBJECTS", "1"); err != nil {
		logger.Error("could not set GIT_NO_REPLACE_OBJECTS=1: %w", err)
	}
}

// GitRepoOptions stores options specific to GitRepo scan requests
type GitRepoOptions struct {
	// Only scan this branch
	Branch string `json:"branch"`
	// Only scan this many commits (reduced if larger than the max scan depth)
	Depth uint16 `json:"depth"`
	// Only scan since this date
	Since string `json:"since"`
	// Work through a proxy for this request
	Proxy string `json:"proxy"`
	// The scan priority
	Priority int `json:"priority"`
}

// GitRepo provides a way to interact with a git repo
type GitRepo struct {
	// Provide common helper functions
	BaseResource

	clonePath    string
	cloneTimeout time.Duration
	cloneURL     string
	options      *GitRepoOptions
}

// NewGitRepo returns a configured git repo resource for the scanner to scan
func NewGitRepo(cloneURL string, options *GitRepoOptions) *GitRepo {
	return &GitRepo{
		cloneURL: cloneURL,
		options:  options,
	}
}

// Kind of resource (always returns GitRepo here
func (r *GitRepo) Kind() string {
	return "GitRepo"
}

// String representation of the resource
func (r *GitRepo) String() string {
	return r.cloneURL
}

// Clone the resource to the desired local location and store the path
func (r *GitRepo) Clone(path string) error {
	r.clonePath = path

	cloneArgs := []string{"clone"}

	if len(r.options.Proxy) > 0 {
		cloneArgs = append(cloneArgs, "--config")
		cloneArgs = append(cloneArgs, fmt.Sprintf("http.proxy=%s", r.options.Proxy))
	}

	// The --[no-]single-branch flags are still needed with mirror due to how
	// things like --depth and --shallow-since behave
	if len(r.options.Branch) > 0 {
		if !r.RemoteRefExists(r.options.Branch) {
			return fmt.Errorf("remote ref does not exist resource_id=%q ref=%q", r.ID(), r.options.Branch)
		}

		cloneArgs = append(cloneArgs, "--bare")
		cloneArgs = append(cloneArgs, "--single-branch")
		cloneArgs = append(cloneArgs, "--branch")
		cloneArgs = append(cloneArgs, r.options.Branch)
	} else {
		cloneArgs = append(cloneArgs, "--mirror")
		cloneArgs = append(cloneArgs, "--no-single-branch")
	}

	if len(r.options.Since) > 0 {
		cloneArgs = append(cloneArgs, "--shallow-since")
		cloneArgs = append(cloneArgs, r.options.Since)
	}

	if r.options.Depth > 0 {
		cloneArgs = append(cloneArgs, "--depth")
		// Add 1 to the clone depth to avoid scanning a grafted commit
		cloneArgs = append(cloneArgs, fmt.Sprint(r.Depth()+1))
	}

	// Include the clone URL
	cloneArgs = append(cloneArgs, r.String(), r.ClonePath())
	var gitClone *exec.Cmd
	var ctx context.Context

	if r.cloneTimeout > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), r.cloneTimeout)
		defer cancel()
		gitClone = exec.CommandContext(ctx, "git", cloneArgs...)
	} else {
		gitClone = exec.Command("git", cloneArgs...)
	}

	output, err := gitClone.CombinedOutput()

	if err != nil {
		return fmt.Errorf("git clone: %w resource_id=%q command=%q output=%q", err, r.ID(), gitClone.String(), output)
	}

	r.Debug(logger.CloneDetail, "git clone resource_id=%q command=%q output=%q", r.ID(), gitClone.String(), output)

	if ctx != nil && ctx.Err() == context.DeadlineExceeded {
		return fmt.Errorf("clone timeout exceeded: %w resource_id=%q", ctx.Err(), r.ID())
	}

	return nil
}

// ClonePath returns where this repo has been cloned if cloned else ""
func (r *GitRepo) ClonePath() string {
	return r.clonePath
}

// EnrichResult enriches the result with contextual information
func (r *GitRepo) EnrichResult(result *response.Result) *response.Result {
	result.Kind = response.GitCommitResultKind
	return result
}

// Branch returns the branch of the repo to scan
func (r *GitRepo) Branch() string {
	return r.options.Branch
}

// Depth returns the depth for things that have version control
func (r *GitRepo) Depth() uint16 {
	return r.options.Depth
}

// SetDepth allows you to adjust the depth for the resource
func (r *GitRepo) SetDepth(depth uint16) {
	r.options.Depth = depth
}

// SetCloneTimeout lets you adjust the timeout before the clone aborts
func (r *GitRepo) SetCloneTimeout(timeout time.Duration) {
	r.cloneTimeout = timeout
}

// Since returns the date after which things should be scanned for things
// that have versions
func (r *GitRepo) Since() string {
	return r.options.Since
}

// ReadFile provides a way to get files out of the repo
func (r *GitRepo) ReadFile(path string) ([]byte, error) {
	object := fmt.Sprintf("HEAD:%s", filepath.Clean(path))
	return exec.Command("git", "-C", r.ClonePath(), "show", object).Output() // #nosec G204
}

// GitDirPath returns the path to the git dir so that other things don't need
// to know how the repo was cloned
func (r *GitRepo) GitDirPath() string {
	// Since --mirror implies --bare, the GitDirPath is the ClonePath
	return r.ClonePath()
}

// ShallowCommits returns a list of shallow commits in a git repo
func (r *GitRepo) ShallowCommits() []string {
	shallowFilePath := filepath.Join(r.GitDirPath(), "shallow")
	data, err := os.ReadFile(filepath.Clean(shallowFilePath))

	var shallowCommits []string

	if err == nil {
		for _, shallowCommit := range strings.Split(string(data), "\n") {
			// Skip empty lines
			if len(shallowCommit) == 0 {
				continue
			}

			shallowCommits = append(shallowCommits, shallowCommit)
		}

		return shallowCommits
	}

	return shallowCommits
}

// Walk traverses the HEAD of a git repo like it's a directory tree. This
// exists this way so even a bare repo can be crawled if needed. To crawl
// different branches, change HEAD.
func (r *GitRepo) Walk(fn WalkFunc) error {
	cmd := exec.Command("git", "-C", r.ClonePath(), "ls-tree", "-r", "--name-only", "--full-tree", "HEAD") // #nosec G204
	output, err := cmd.Output()

	if err != nil {
		return fmt.Errorf("could not list files: %w", err)
	}

	for _, path := range strings.Split(string(output), "\n") {
		if len(path) == 0 {
			continue
		}

		data, err := r.ReadFile(path)
		if err == nil {
			return err
		}

		if err := fn(path, bytes.NewReader(data)); err != nil {
			return err
		}
	}

	return nil
}

// RemoteRefExists checks the remote repository to see if the ref exists
func (r *GitRepo) RemoteRefExists(ref string) bool {
	cmd := exec.Command("git", "ls-remote", "--exit-code", "--quiet", r.String(), ref) // #nosec G204
	return cmd.Run() == nil
}

// Refs returns the unique OIDs in a repo
func (r *GitRepo) Refs() []string {
	cmd := exec.Command("git", "-C", r.ClonePath(), "show-ref", "--hash") // #nosec G204
	out, err := cmd.Output()
	refs := []string{}

	if err != nil {
		r.Error(logger.CommandError, "could not list refs: %w", err)
		return refs
	}

	for _, ref := range bytes.Split(out, []byte{'\n'}) {
		refStr := strings.TrimSpace(string(ref))

		if len(refStr) == 0 {
			continue
		}

		if !slices.Contains(refs, refStr) {
			refs = append(refs, refStr)
		}
	}

	return refs
}

// Priority returns the scan priority
func (r *GitRepo) Priority() int {
	return r.options.Priority
}
