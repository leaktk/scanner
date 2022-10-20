use serde::Deserialize;
use std::path::Path;

pub const SCANNER: &str = "gitleaks";
pub const VERSION: &str = "8.12.0";
pub const PATTERNS_FILE: &str = "patterns.toml";

#[derive(Debug, Deserialize)]
pub struct PatternsConfig {
    pub server_url: String,
    pub refresh_interval: u64,
}

#[derive(Debug, Deserialize)]
pub enum ListnerMethod {
    #[serde(rename = "stdin")]
    Stdin,
}

#[derive(Debug, Deserialize)]
pub struct ListnerConfig {
    pub method: ListnerMethod,
}

#[derive(Debug, Deserialize)]
pub struct ScannerConfig {
    pub workdir: Box<Path>,
    pub patterns: PatternsConfig,
}

#[derive(Debug, Deserialize)]
pub struct Config {
    pub listner: ListnerConfig,
    pub scanner: ScannerConfig,
}

impl Config {
    pub fn load(raw: &str) -> Config {
        toml::from_str(raw).expect("Could not load config file")
    }
}
