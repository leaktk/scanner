use crate::config::ScannerConfig;
use crate::errors::Error;
use log::{debug, info};
use std::fs::{self, File};
use std::io::Write;
use std::path::PathBuf;
use std::time::{self, Duration, SystemTime};

pub struct Patterns {
    pub path: PathBuf,
    refresh_interval: Duration,
    server_url: String,
    gitleaks_version: String,
}

impl Patterns {
    pub fn new(config: &ScannerConfig) -> Patterns {
        Patterns {
            refresh_interval: Duration::from_secs(config.patterns.refresh_interval),
            server_url: config.patterns.server_url.clone(),
            gitleaks_version: config.gitleaks.version.clone(),
            path: config.workdir.join("patterns").join(format!(
                "gitleaks-{}-patterns.toml",
                config.gitleaks.version
            )),
        }
    }

    // Block and refresh the patterns file if it needs to be refreshed
    pub fn refresh_if_stale(&self) -> Result<(), Error> {
        if !self.are_stale() {
            debug!("Patterns up-to-date");
            Ok(())
        } else {
            info!("Refreshing patterns");

            let url = format!(
                "{}/patterns/gitleaks/{}",
                self.server_url, self.gitleaks_version,
            );

            fs::create_dir_all(self.path.parent().expect("patterns in a directory"))?;

            let content = reqwest::blocking::get(url)?.bytes()?;
            let mut file = File::create(&self.path)?;

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
        self.path
            .metadata()
            .and_then(|m| m.modified())
            .unwrap_or(time::UNIX_EPOCH)
    }
}
