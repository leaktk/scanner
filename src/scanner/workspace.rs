use log::error;
use std::fmt;
use std::fs;
use std::path::{Path, PathBuf};

//
pub struct Workspace {
    root_dir: PathBuf,

    // Point the scanners at this directory
    pub scan_dir: PathBuf,

    // Save repo specific configs in this directory
    pub config_dir: PathBuf,

    // Save the results in this directory
    pub results_dir: PathBuf,
}

impl Workspace {
    pub fn new(root_dir: &Path, external_scan_dir: Option<&Path>) -> Self {
        Workspace {
            root_dir: root_dir.to_path_buf(),
            // A scan dir outside of the root_dir is not cleaned up.
            // This can be useful for local scans where you don't want
            // to delete the local repo
            scan_dir: external_scan_dir
                .unwrap_or(&root_dir.join("repo"))
                .to_path_buf(),
            config_dir: root_dir.join("config"),
            results_dir: root_dir.join("results"),
        }
    }

    pub fn clean(&self) {
        if self.exists() {
            fs::remove_dir_all(&self.root_dir).unwrap_or_else(|err| {
                error!("could not remove {}: {}", self, err);
            });
        }
    }

    pub fn exists(&self) -> bool {
        self.root_dir.exists()
    }
}

impl fmt::Display for Workspace {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        write!(f, "{}", self.root_dir.display())
    }
}
