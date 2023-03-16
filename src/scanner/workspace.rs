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
}

impl Workspace {
    pub fn new(root_dir: &Path) -> Self {
        Workspace {
            root_dir: root_dir.to_path_buf(),
            scan_dir: root_dir.join("repo"),
            config_dir: root_dir.join("config"),
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
