use serde::Deserialize;

pub const SCANNER: &str = "gitleaks";
pub const VERSION: &str = "7.6.1";
pub const PATTERNS_FILE: &str = "patterns.toml";

#[derive(Debug, Deserialize)]
pub struct Patterns {
    pub server_url: String,
    pub refresh_interval: u32,
}

#[derive(Debug, Deserialize)]
pub struct Scanner {
    pub workdir: String,
    pub patterns: Patterns,
}

#[derive(Debug, Deserialize)]
pub struct Config {
    pub scanner: Scanner
}

impl Config {
    pub fn load(raw: &str) -> Config {
        toml::from_str(raw).expect("Could not load config file")
    }
}
