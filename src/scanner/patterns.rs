use std::fs::{self, File};
use std::io::Write;
use std::path::PathBuf;
use std::time::{self, Duration, SystemTime};

use log::{debug, info};
use reqwest::header;
use thiserror::Error;

use crate::config::ScannerConfig;

#[derive(Error, Debug)]
pub enum PatternsError {
    #[error("could not fetch patterns: {0}")]
    CouldNotFetch(String),

    #[error("could not save patterns: {0}")]
    CouldNotSave(#[from] std::io::Error),

    #[error("auth token is invalid: {0}")]
    InvalidAuthToken(#[from] reqwest::header::InvalidHeaderValue),

    #[error("patterns path has no parent")]
    PathHasNoParent,
}

impl From<reqwest::Error> for PatternsError {
    fn from(error: reqwest::Error) -> Self {
        PatternsError::CouldNotFetch(error.to_string())
    }
}

struct PatternServer {
    url: String,
    auth_token: Option<String>,
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
                auth_token: config.patterns.server.auth_token.clone(),
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

            let mut headers = header::HeaderMap::new();

            if let Some(auth_token) = &self.server.auth_token {
                headers.insert(
                    header::AUTHORIZATION,
                    header::HeaderValue::from_str(&format!("Bearer {}", auth_token))?,
                );
            }

            let client = reqwest::blocking::Client::builder()
                .default_headers(headers)
                .build()?;

            let url = self.server.gitleaks_patterns_url();
            let resp = client.get(url).send()?;

            if !resp.status().is_success() {
                return Err(PatternsError::CouldNotFetch(resp.status().to_string()));
            }

            let content = resp.bytes()?;

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
