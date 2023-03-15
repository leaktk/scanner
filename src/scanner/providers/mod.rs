mod git;

use std::path::Path;

use thiserror::Error;

use super::proto::{Request, RequestKind};

use git::{Git, GitError};

#[derive(Error, Debug)]
pub enum ProviderError {
    #[error(transparent)]
    GitCouldNotClone(#[from] GitError),
}

pub struct Providers {
    pub git: Git,
}

// This handles building all of the providers for the scanner to use and
impl Providers {
    pub fn new() -> Providers {
        Providers { git: Git::new() }
    }

    pub fn clone(&self, req: &Request, dest: &Path) -> Result<String, ProviderError> {
        if req.is_local() {
            return Ok("skipped clone for local target".to_string());
        }

        match req.kind {
            RequestKind::Git => {
                let output = self.git.clone(&req.target, &req.options, &dest)?;
                Ok(String::from_utf8_lossy(&output.stderr).to_string())
            }
        }
    }
}
