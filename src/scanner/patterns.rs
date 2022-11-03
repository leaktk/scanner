use crate::config::{ScannerConfig, GITLEAKS_VERSION, PATTERNS_FILE, SCANNER};
use std::fs::{self, File};
use std::io::Write;
use std::path::PathBuf;

// TODO: clean this up after learning more about rust matching

fn write_patterns_file(config: &ScannerConfig, patterns: &str) {
    let patterns_dir = patterns_dir(&config);
    fs::create_dir_all(&patterns_dir).expect("Could not create patterns file directory!");

    match File::create(patterns_path(&config)) {
        Ok(mut patterns_file) => {
            match patterns_file.write_all(patterns.as_bytes()) {
                // TODO: replace with log function
                Ok(_) => println!("Patterns updated!"),
                // TODO: replace with log function
                Err(e) => println!("{:#?}", e),
            }
        }
        // TODO: replace with log function
        Err(e) => println!("{:#?}", e),
    }
}

// Block and refresh the patterns file
// If there is an error and a patterns file already exists: just log it
// If there is an error and a patterns file does not exist: panic
pub fn refresh(config: &ScannerConfig) {
    let url = format!(
        "{}/{}/{}",
        config.patterns.server_url, SCANNER, GITLEAKS_VERSION
    );

    match reqwest::blocking::get(url) {
        Ok(resp) => {
            match resp.text() {
                Ok(body) => write_patterns_file(&config, &body),
                // TODO: replace with log function
                Err(e) => println!("{:#?}", e),
            }
        }
        // TODO: replace with log function
        Err(_) => println!("There was an error updating the patterns file!"),
    }

    // TODO: panic if the patterns file doesn't exist
}

fn patterns_dir(config: &ScannerConfig) -> PathBuf {
    config.workdir.join("patterns").join(GITLEAKS_VERSION)
}

pub fn patterns_path(config: &ScannerConfig) -> PathBuf {
    patterns_dir(config).join(PATTERNS_FILE)
}
