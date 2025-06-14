package resource

//
// import (
// 	"context"
// 	"fmt"
// 	"os/exec"
// 	"time"
//
// 	"github.com/leaktk/leaktk/pkg/logger"
// )
//
// // GitRepoOptions stores options specific to GitRepo scan requests
// type GitCloneOptions struct {
// 	Branch   string `json:"branch"`
// 	Depth    int    `json:"depth"`
// 	Local    bool   `json:"local"`
// 	Staged   bool   `json:"staged"`
// 	Since    string `json:"since"`
// 	Proxy    string `json:"proxy"`
// 	Priority int    `json:"priority"`
// 	Unstaged bool   `json:"unstaged"`
// }
//
// // GitRepo provides a way to interact with a git repo
// type GitRepo struct {
// 	path         string
// 	cloneTimeout time.Duration
// 	repo         string
// 	options      *GitRepoOptions
// }
//
// // Clone the resource to the desired local location and store the path
// func (r *GitRepo) Clone(path string) error {
// 	if r.path != "" {
// 		return fmt.Errorf("resource path already set: path=%q", r.path)
// 	}
//
// 	r.path = path
//
// 	cloneArgs := []string{"clone"}
//
// 	if len(r.options.Proxy) > 0 {
// 		cloneArgs = append(cloneArgs, "--config")
// 		cloneArgs = append(cloneArgs, fmt.Sprintf("http.proxy=%s", r.options.Proxy))
// 	}
//
// 	// The --[no-]single-branch flags are still needed with mirror due to how
// 	// things like --depth and --shallow-since behave
// 	if len(r.options.Branch) > 0 {
// 		if !r.RemoteRefExists(r.options.Branch) {
// 			return fmt.Errorf("remote ref does not exist: resource_id=%q ref=%q", r.ID(), r.options.Branch)
// 		}
//
// 		cloneArgs = append(cloneArgs, "--bare")
// 		cloneArgs = append(cloneArgs, "--single-branch")
// 		cloneArgs = append(cloneArgs, "--branch")
// 		cloneArgs = append(cloneArgs, r.options.Branch)
// 	} else {
// 		cloneArgs = append(cloneArgs, "--mirror")
// 		cloneArgs = append(cloneArgs, "--no-single-branch")
// 	}
//
// 	if len(r.options.Since) > 0 {
// 		cloneArgs = append(cloneArgs, "--shallow-since")
// 		cloneArgs = append(cloneArgs, r.options.Since)
// 	}
//
// 	if r.options.Depth > 0 {
// 		cloneArgs = append(cloneArgs, "--depth")
// 		// Add 1 to the clone depth to avoid scanning a grafted commit
// 		cloneArgs = append(cloneArgs, fmt.Sprint(r.Depth()+1))
// 	}
//
// 	// Include the clone URL
// 	cloneArgs = append(cloneArgs, r.String(), r.Path())
// 	var gitClone *exec.Cmd
// 	var ctx context.Context
//
// 	if r.cloneTimeout > 0 {
// 		ctx, cancel := context.WithTimeout(context.Background(), r.cloneTimeout)
// 		defer cancel()
// 		gitClone = exec.CommandContext(ctx, "git", cloneArgs...)
// 	} else {
// 		gitClone = exec.Command("git", cloneArgs...)
// 	}
//
// 	output, err := gitClone.CombinedOutput()
//
// 	if err != nil {
// 		return fmt.Errorf("git clone: resource_id=%q command=%q error=%q output=%q", r.ID(), gitClone.String(), err.Error(), output)
// 	}
//
// 	r.Debug(logger.CloneDetail, "git clone: resource_id=%q command=%q output=%q", r.ID(), gitClone.String(), output)
//
// 	if ctx != nil && ctx.Err() == context.DeadlineExceeded {
// 		return fmt.Errorf("clone timeout exceeded resource_id=%q error=%q", r.ID(), ctx.Err().Error())
// 	}
//
// 	return nil
// }
//
// // RemoteRefExists checks the remote repo to see if the ref exists
// func (r *GitRepo) RemoteRefExists(ref string) bool {
// 	cmd := exec.Command("git", "ls-remote", "--exit-code", "--quiet", r.String(), ref) // #nosec G204
// 	return cmd.Run() == nil
// }
