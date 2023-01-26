mod git;

use super::proto::Request;
use git::Git;
use std::io::Error;
use std::path::Path;
use std::process::Output;

pub struct Providers {
    pub git: Git,
}

// This handles building all of the providers for the scanner to use and
impl Providers {
    pub fn new() -> Providers {
        Providers { git: Git::new() }
    }

    pub fn clone(&self, req: &Request, dest: &Path) -> Result<Output, Error> {
        match req {
            Request::Git { url, options, .. } => self.git.clone(&url, &options, &dest),
        }
    }
}
