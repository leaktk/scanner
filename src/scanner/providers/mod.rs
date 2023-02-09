mod git;

use super::proto::{Request, RequestKind};
use git::Git;
use std::path::Path;

pub struct Providers {
    pub git: Git,
}

pub struct CloneResult {
    pub ok: bool,
    pub msg: String,
}

// This handles building all of the providers for the scanner to use and
impl Providers {
    pub fn new() -> Providers {
        Providers { git: Git::new() }
    }

    pub fn clone(&self, req: &Request, dest: &Path) -> CloneResult {
        match req.kind {
            RequestKind::Git => {
                if req.is_local() {
                    CloneResult {
                        ok: true,
                        msg: "Skipped clone for local repo".to_string(),
                    }
                } else {
                    match self.git.clone(&req.target, &req.options, &dest) {
                        Err(err) => CloneResult {
                            ok: false,
                            msg: err.to_string(),
                        },
                        Ok(output) => CloneResult {
                            ok: output.status.success(),
                            msg: String::from_utf8_lossy(&output.stderr).to_string(),
                        },
                    }
                }
            }
        }
    }
}
