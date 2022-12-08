use crate::scanner::proto::GitOptions;
use std::path::Path;
use std::process::Command;

pub fn clone(clone_url: &String, options: &Option<GitOptions>, clone_dir: &Path) {
    // TODO: logging
    let mut git = Command::new("git");

    let git_clone = match &options {
        None => git.arg("clone").arg(&clone_url).arg(clone_dir),
        Some(opts) => {
            let mut args = Vec::new();

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
                args.push(format!("--depth={depth}"));
            }

            if let Some(shallow_since) = &opts.shallow_since {
                args.push(format!("--shallow-since={shallow_since}"));
            }

            git.arg("clone").args(args).arg(&clone_url).arg(clone_dir)
        }
    };

    // TODO: Handle errors
    git_clone.output().expect("Could not clone repo!");
}
