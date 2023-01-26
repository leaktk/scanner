use crate::scanner::proto::GitOptions;
use log::info;
use std::fs;
use std::io::Error;
use std::path::Path;
use std::process::{Command, Output};

pub struct Git;

impl Git {
    pub fn new() -> Git {
        // Created for when default options, etc might be added
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
        options: &Option<GitOptions>,
        clone_dir: &Path,
    ) -> Result<Output, Error> {
        let mut args: Vec<String> = Vec::new();

        if let Some(opts) = &options {
            if let Some(configs) = &opts.config {
                for config in configs {
                    args.push(format!("--config={config}"));
                }
            }

            if let Some(branch) = &opts.branch {
                args.push(format!("--branch={branch}"));
            }

            if let Some(single_branch) = opts.single_branch {
                if single_branch {
                    args.push("--single-branch".to_string());
                } else {
                    args.push("--no-single-branch".to_string());
                }
            }

            if let Some(depth) = opts.depth {
                // Add one to the depth since the grafted commit is excluded
                // from the scan
                let depth = depth + 1;
                args.push(format!("--depth={depth}"));
            }

            if let Some(shallow_since) = &opts.shallow_since {
                args.push(format!("--shallow-since={shallow_since}"));
            }
        }

        info!(
            "Running: git clone {} {} {}",
            args.join(" "),
            clone_url,
            clone_dir.display()
        );

        Command::new("git")
            .arg("clone")
            .args(args)
            .arg(clone_url)
            .arg(clone_dir)
            .env("GIT_TERMINAL_PROMPT", "0")
            .output()
    }
}
