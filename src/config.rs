use crate::errors::Error;
use serde::Deserialize;
use std::fs;
use std::path::Path;

pub const SCANNER: &str = "gitleaks";
pub const PATTERNS_FILE: &str = "patterns.toml";

// TODO: break this into some OS aware object or put it into the config file
pub const GITLEAKS_VERSION: &str = "8.12.0";
pub const GITLEAKS_LINUX_X64_URL: &str =
    "https://raw.githubusercontent.com/leaktk/bin/main/bin/gitleaks-8.12.0-linux-x86_64";
pub const GITLEAKS_LINUX_X64_CHECKSUM: &str =
    "9ed4271ffbfa04feec1423eb56154ad22c11ac4cae698f0115c1a064d4553524";

#[derive(Debug, Deserialize)]
pub struct PatternsConfig {
    pub server_url: String,
    pub refresh_interval: u64,
}

#[derive(Debug, Deserialize)]
pub struct LoggerConfig {
    #[serde(default = "LoggerConfig::default_level")]
    pub level: log::Level,
}

#[derive(Debug, Deserialize)]
pub struct ScannerConfig {
    pub workdir: Box<Path>,
    pub patterns: PatternsConfig,
}

#[derive(Debug, Deserialize)]
pub struct Config {
    #[serde(default)]
    pub logger: LoggerConfig,
    pub scanner: ScannerConfig,
}

impl LoggerConfig {
    // Define the default level for the LoggerConfig if the section
    // is specified but the level attr isn't
    fn default_level() -> log::Level {
        log::Level::Info
    }
}

// Define what a default LoggerConfig should be if the section isn't specified
impl Default for LoggerConfig {
    fn default() -> Self {
        LoggerConfig {
            level: log::Level::Info,
        }
    }
}

impl Config {
    pub fn from_str(raw: &str) -> Result<Config, Error> {
        toml::from_str(raw).map_err(|err| Error::new(err.to_string()))
    }

    // Load the config from a file path
    pub fn load(path: &str) -> Result<Config, Error> {
        let content = fs::read_to_string(path)
            .map_err(|err| Error::new(format!("Could not read {}: {}", path, err)))?;

        Config::from_str(&content)
    }
}
