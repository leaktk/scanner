mod git;

use git::Git;

pub struct Providers {
    pub git: Git,
}

// This handles building all of the providers for the scanner to use and
impl Providers {
    pub fn new() -> Providers {
        Providers { git: Git::new() }
    }
}
