use crate::config::{ScannerConfig, GITLEAKS_VERSION, PATTERNS_FILE, SCANNER};
use log::{error, info};
use std::fs::{self, File};
use std::io::Write;
use std::path::PathBuf;

// TODO: clean this up after learning more about rust matching

fn write_patterns_file(config: &ScannerConfig, patterns: &str) {
    let patterns_dir = patterns_dir(&config);
    fs::create_dir_all(&patterns_dir).expect("Could not create patterns file directory!");

    match File::create(patterns_path(&config)) {
        Ok(mut patterns_file) => match patterns_file.write_all(patterns.as_bytes()) {
            Ok(_) => info!("Patterns updated!"),
            Err(e) => error!("{:#?}", e),
        },
        Err(e) => error!("{:#?}", e),
    }
}

// Block and refresh the patterns file
// If there is an error and a patterns file already exists: just log it
// If there is an error and a patterns file does not exist: panic
pub fn refresh(config: &ScannerConfig) {
    let url = format!(
        "{}/patterns/{}/{}",
        config.patterns.server_url, SCANNER, GITLEAKS_VERSION
    );

    match reqwest::blocking::get(url) {
        Ok(resp) => match resp.text() {
            Ok(body) => write_patterns_file(&config, &body),
            Err(e) => error!("{:#?}", e),
        },
        Err(_) => error!("There was an error updating the patterns file!"),
    }

    // TODO: panic if the patterns file doesn't exist
}

fn patterns_dir(config: &ScannerConfig) -> PathBuf {
    config.workdir.join("patterns").join(GITLEAKS_VERSION)
}

pub fn patterns_path(config: &ScannerConfig) -> PathBuf {
    patterns_dir(config).join(PATTERNS_FILE)
}
