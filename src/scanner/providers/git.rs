use crate::scanner::proto::GitOptions;
use std::path::Path;
use std::process::Command;

pub fn clone(id: &String, clone_url: &String, options: &Option<GitOptions>, clone_dir: &Path) {
    // TODO: logging
    let mut git = Command::new("git");

    let git_clone = match &options {
        None => git.arg("clone").arg(&clone_url).arg(clone_dir),
        Some(opts) => {
            let mut args = Vec::new();

            if let Some(clone_depth) = opts.clone_depth {
                // TODO: sanitize input
                args.push(format!("--depth={clone_depth}"));
            }

            // TODO: Add additional options here

            git.arg("clone").args(args).arg(&clone_url).arg(clone_dir)
        }
    };

    // TODO: Handle errors
    git_clone.output().expect("Could not clone repo!");
}
