package resource

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/leaktk/scanner/pkg/logger"
)

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

	cloneArgs := []string{"clone"}

	// Add expanded refs to capture things like PRs etc.
	cloneArgs = append(cloneArgs, "--config")
	cloneArgs = append(cloneArgs, "remote.origin.fetch=+refs/*:refs/remotes/origin/*")

	if len(r.options.Proxy) > 0 {
		cloneArgs = append(cloneArgs, "--config")
		cloneArgs = append(cloneArgs, fmt.Sprintf("http.proxy=%s", r.options.Proxy))
	}

	if len(r.options.Since) > 0 {
		cloneArgs = append(cloneArgs, "--shallow-since")
		cloneArgs = append(cloneArgs, r.options.Since)
	}

	if len(r.options.Branch) > 0 {
		cloneArgs = append(cloneArgs, "--single-branch")
		cloneArgs = append(cloneArgs, "--branch")
		cloneArgs = append(cloneArgs, r.options.Branch)
	} else {
		cloneArgs = append(cloneArgs, "--no-single-branch")
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
