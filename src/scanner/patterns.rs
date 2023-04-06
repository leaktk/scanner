use std::fs::{self, File};
use std::io::Write;
use std::path::PathBuf;
use std::time::{self, Duration, SystemTime};

use log::{debug, info};
use thiserror::Error;

use crate::config::ScannerConfig;

#[derive(Error, Debug)]
pub enum PatternsError {
    #[error("could not fetch patterns")]
    CouldNotFetch(#[from] reqwest::Error),

    #[error("could not save patterns")]
    CouldNotSave(#[from] std::io::Error),

    #[error("patterns path has no parent")]
    PathHasNoParent,
}

struct PatternServer {
    url: String,
    gitleaks_version: String,
}

impl PatternServer {
    fn gitleaks_patterns_url(&self) -> String {
        format!("{}/patterns/gitleaks/{}", self.url, self.gitleaks_version)
    }
}

pub struct Patterns {
    pub gitleaks_patterns_path: PathBuf,
    refresh_interval: Duration,
    server: PatternServer,
}

impl Patterns {
    pub fn new(config: &ScannerConfig) -> Patterns {
        Patterns {
            refresh_interval: Duration::from_secs(config.patterns.refresh_interval),
            server: PatternServer {
                url: config.patterns.server.url.clone(),
                gitleaks_version: config.gitleaks.version.clone(),
            },
            gitleaks_patterns_path: config.workdir.join("patterns").join(format!(
                "gitleaks-{}-patterns.toml",
                config.gitleaks.version
            )),
        }
    }

    #[inline]
    pub fn exists(&self) -> bool {
        self.gitleaks_patterns_path.exists()
    }

    // Block and refresh the patterns file if it needs to be refreshed
    pub fn refresh_if_stale(&self) -> Result<(), PatternsError> {
        if !self.are_stale() {
            debug!("patterns are already up-to-date");
            Ok(())
        } else {
            info!("refreshing patterns");

            fs::create_dir_all(
                self.gitleaks_patterns_path
                    .parent()
                    .ok_or(PatternsError::PathHasNoParent)?,
            )?;

            let url = self.server.gitleaks_patterns_url();
            let content = reqwest::blocking::get(url)?.bytes()?;

            File::create(&self.gitleaks_patterns_path)?.write_all(&content)?;
            info!("patterns refreshed");

            Ok(())
        }
    }

    #[inline]
    fn are_stale(&self) -> bool {
        SystemTime::now()
            .duration_since(self.modified())
            .map_or(true, |d| d > self.refresh_interval)
    }

    #[inline]
    fn modified(&self) -> SystemTime {
        self.gitleaks_patterns_path
            .metadata()
            .and_then(|m| m.modified())
            .unwrap_or(time::UNIX_EPOCH)
    }
}
