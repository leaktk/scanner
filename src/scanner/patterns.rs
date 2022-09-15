use crate::config::{ScannerConfig, PATTERNS_FILE, SCANNER, VERSION};
use std::fs::{self, File};
use std::io::Write;
use std::path::Path;

// TODO: clean this up after learning more about rust matching

fn write_patterns_file(workdir: &Path, patterns: &str) {
    let patterns_dir = workdir.join("patterns").join(VERSION);
    fs::create_dir_all(&patterns_dir).expect("Could not create patterns file directory!");

    match File::create(&patterns_dir.join(PATTERNS_FILE)) {
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
    let url = format!("{}/{}/{}", config.patterns.server_url, SCANNER, VERSION);

    match reqwest::blocking::get(url) {
        Ok(resp) => {
            match resp.text() {
                Ok(body) => write_patterns_file(&config.workdir, &body),
                // TODO: replace with log function
                Err(e) => println!("{:#?}", e),
            }
        }
        // TODO: replace with log function
        Err(_) => println!("There was an error updating the patterns file!"),
    }

    // TODO: panic if the patterns file doesn't exist
}
