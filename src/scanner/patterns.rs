use crate::config::ScannerConfig;
use crate::errors::Error;
use log::info;
use std::fs::{self, File};
use std::io::Write;
use std::path::PathBuf;

// Block and refresh the patterns file
pub fn refresh(config: &ScannerConfig) -> Result<(), Error> {
    info!("Refreshing patterns");

    let url = format!(
        "{}/patterns/gitleaks/{}",
        config.patterns.server_url, config.gitleaks.version,
    );

    fs::create_dir_all(&patterns_dir(&config))?;

    let content = reqwest::blocking::get(url)?.bytes()?;
    let mut file = File::create(patterns_path(&config))?;

    file.write_all(&content)?;

    info!("Patterns refreshed!");
    Ok(())
}

fn patterns_dir(config: &ScannerConfig) -> PathBuf {
    config
        .workdir
        .join("patterns")
        .join(&config.gitleaks.version)
}

pub fn patterns_path(config: &ScannerConfig) -> PathBuf {
    patterns_dir(config).join(format!(
        "gitleaks-{}-patterns.toml",
        config.gitleaks.version
    ))
}
