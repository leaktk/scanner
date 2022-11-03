use serde::Deserialize;
use std::path::Path;

pub const SCANNER: &str = "gitleaks";
pub const PATTERNS_FILE: &str = "patterns.toml";

// TODO: break this into some OS aware object or put it into the config file
pub const GITLEAKS_VERSION: &str = "8.12.0";
pub const GITLEAKS_LINUX_X64_URL: &str = "https://raw.githubusercontent.com/leaktk/bin/main/bin/gitleaks-8.12.0-linux-x86_64";
pub const GITLEAKS_LINUX_X64_CHECKSUM: &str = "9ed4271ffbfa04feec1423eb56154ad22c11ac4cae698f0115c1a064d4553524";

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
