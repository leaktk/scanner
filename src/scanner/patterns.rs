use crate::config::ScannerConfig;
use crate::errors::Error;
use log::{debug, info};
use std::fs::{self, File};
use std::io::Write;
use std::path::PathBuf;
use std::time::{self, Duration, SystemTime};

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
    pub fn refresh_if_stale(&self) -> Result<(), Error> {
        if !self.are_stale() {
            debug!("Patterns up-to-date");
            Ok(())
        } else {
            info!("Refreshing patterns");

            fs::create_dir_all(
                self.gitleaks_patterns_path
                    .parent()
                    .expect("patterns in a directory"),
            )?;

            let url = self.server.gitleaks_patterns_url();
            let content = reqwest::blocking::get(url)?.bytes()?;
            let mut file = File::create(&self.gitleaks_patterns_path)?;

            file.write_all(&content)?;

            info!("Patterns refreshed!");
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
