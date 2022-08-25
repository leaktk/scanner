use crate::config::{Scanner, SCANNER, VERSION, PATTERNS_FILE};
use std::fs::{self, File};
use std::io::Write;


// TODO: clean this up after I learn more about rust matching

fn write_patterns_file(path: &str, patterns: &str) {
    fs::create_dir_all(path).expect("Could not create patterns file directory!");

    // TODO: make this portable
    let patterns_path = format!("{}/{}", path, PATTERNS_FILE);

    match File::create(patterns_path) {
        Ok(mut patterns_file) => {
            match patterns_file.write_all(patterns.as_bytes()) {
                // TODO: replace with log function
                Ok(_) => println!("Patterns updated!"),
                // TODO: replace with log function
                Err(e) => println!("{:#?}", e),
            }
        },
        // TODO: replace with log function
        Err(e) => println!("{:#?}", e),
    }
}

// Block and refresh the patterns file
// If there is an error and a patterns file already exists: just log it
// If there is an error and a patterns file does not exist: panic
pub fn refresh(scanner: &Scanner) {
    let url = format!("{}/{}/{}", scanner.patterns.server_url, SCANNER, VERSION);

    match reqwest::blocking::get(url) {
        Ok(resp) => {
            match resp.text() {
                Ok(body) => write_patterns_file(&scanner.workdir, &body),
                // TODO: replace with log function
                Err(e) => println!("{:#?}", e),
            }
        },
        // TODO: replace with log function
        Err(_) => println!("There was an error updating the patterns file!"),
    }

    // TODO: panic if the patterns file doesn't exist
}
