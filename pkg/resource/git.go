package resource

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/leaktk/scanner/pkg/logger"
)

// Configure git env
func init() {
	// Make sure git never prompts for a password
	if err := os.Setenv("GIT_TERMINAL_PROMPT", "0"); err != nil {
		logger.Error("could not set GIT_TERMINAL_PROMPT=0: error=%q", err)
	}

	// Make sure replace is disabled so we can scan all of the refs
	if err := os.Setenv("GIT_NO_REPLACE_OBJECTS", "1"); err != nil {
		logger.Error("could not set GIT_NO_REPLACE_OBJECTS=1: error=%q", err)
	}
}

// GitRepoOptions stores options specific to GitRepo scan requests
type GitRepoOptions struct {
	// Only scan this many commits (reduced if larger than the max scan depth)
	Depth uint16 `json:"depth"`
	// Only scan since this date
	Since string `json:"since"`
	// Only scan this branch
	Branch string `json:"branch"`
	// Work through a proxy for this request
	Proxy string `json:"proxy"`
}

// GitRepo provides a way to interact with a git repo
type GitRepo struct {
	// Provide common helper functions
	BaseResource

	clonePath    string
	cloneURL     string
	options      *GitRepoOptions
	cloneTimeout time.Duration
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

	cloneArgs := []string{"clone", "--mirror"}

	if len(r.options.Proxy) > 0 {
		cloneArgs = append(cloneArgs, "--config")
		cloneArgs = append(cloneArgs, fmt.Sprintf("http.proxy=%s", r.options.Proxy))
	}

	// The --[no-]single-branch flags are still needed with mirror due to how
	// things like --depth and --shallow-since behave
	if len(r.options.Branch) > 0 {
		cloneArgs = append(cloneArgs, "--single-branch")
		cloneArgs = append(cloneArgs, "--branch")
		cloneArgs = append(cloneArgs, r.options.Branch)
	} else {
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
		return fmt.Errorf("git clone: resource_id=%q command=%q error=%q output=%q", r.ID(), gitClone.String(), err.Error(), output)
	}

	logger.Debug("git clone: resource_id=%q command=%q output=%q", r.ID(), gitClone.String(), output)

	if ctx != nil && ctx.Err() == context.DeadlineExceeded {
		return fmt.Errorf("clone timeout exceeded resource_id=%q error=%q", r.ID(), ctx.Err().Error())
	}

	return nil
}

// ClonePath returns where this repo has been cloned if cloned else ""
func (r *GitRepo) ClonePath() string {
	return r.clonePath
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

// GitDirPath returns the path to the git dir so that other things don'g need
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

// Walk the files in HEAD
func (r *GitRepo) Walk(fn WalkFunc) error {
	cmd := exec.Command("git", "-C", r.ClonePath(), "ls-tree", "-r", "--name-only", "--full-tree", "HEAD") // #nosec G204
	output, err := cmd.Output()

	if err != nil {
		return fmt.Errorf("could not list files: %q", err)
	}

	for _, path := range strings.Split(string(output), "\n") {
		if len(path) == 0 {
			continue
		}

		data, err := r.ReadFile(path)
		if err == nil {
			return err
		}

		if err := fn(path, data); err != nil {
			return err
		}
	}

	return nil
}
