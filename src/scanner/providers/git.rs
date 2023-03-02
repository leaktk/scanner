use std::fs;
use std::path::Path;
use std::process::{Command, Output};

use log::{error, info};
use thiserror::Error;

use crate::scanner::proto::RequestOptions;

#[derive(Error, Debug)]
pub enum GitError {
    #[error("clone failed")]
    CloneFailed(#[from] std::io::Error),

    #[error("invalid clone url")]
    InvalidCloneURL,
}

pub struct Git;

impl Git {
    pub fn new() -> Git {
        Git {}
    }

    // These commits are grafted and should not be included in scans
    pub fn shallow_commits(&self, clone_dir: &Path) -> Vec<String> {
        let shallow_file_path = clone_dir.join(".git").join("shallow");

        if let Ok(shallow_commits) = fs::read_to_string(shallow_file_path) {
            shallow_commits.lines().map(|s| s.to_string()).collect()
        } else {
            vec![]
        }
    }

    pub fn clone(
        &self,
        clone_url: &String,
        options: &RequestOptions,
        clone_dir: &Path,
    ) -> Result<Output, GitError> {
        // This is added here for safety. If this needs to change for some
        // reason, additional sanitation might be needed.
        if !clone_url.starts_with("https://") {
            error!("only https clone urls are supported");
            return Err(GitError::InvalidCloneURL);
        }

        let mut args = vec!["clone".to_string()];

        if let Some(configs) = &options.config {
            for config in configs {
                args.push(format!("--config={config}"));
            }
        }

        if let Some(branch) = &options.branch {
            args.push(format!("--branch={branch}"));
        }

        if let Some(single_branch) = options.single_branch {
            if single_branch {
                args.push("--single-branch".to_string());
            } else {
                args.push("--no-single-branch".to_string());
            }
        }

        if let Some(depth) = options.depth {
            // increment the depth since grafted commits are excluded
            let depth = depth + 1;
            args.push(format!("--depth={depth}"));
        }

        if let Some(since) = &options.since {
            args.push(format!("--shallow-since={since}"));
        }

        args.push(clone_url.to_string());
        args.push(clone_dir.display().to_string());

        info!("running git '{}'", args.join("' '"));

        let output = Command::new("git")
            .args(args)
            .env("GIT_TERMINAL_PROMPT", "0")
            .output()?;

        Ok(output)
    }
}
