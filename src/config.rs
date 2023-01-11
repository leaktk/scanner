use crate::errors::Error;
use serde::Deserialize;
use std::env;
use std::fs;
use std::path::Path;

#[derive(Debug, Deserialize)]
pub struct GitleaksConfig {
    pub filename: String,
    pub version: String,
    pub download_url: String,
    pub checksum: String,
}

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
    #[serde(default)]
    pub gitleaks: GitleaksConfig,
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

// Generally you shouldn't override the gitleaks config but this provides some flexibility and
// provides a consistent way to access the details in an OS agnoistic way. Just make sure it's
// compatible with the version defined in the source.
impl Default for GitleaksConfig {
    fn default() -> Self {
        let version = "8.12.0";

        let (filename, checksum) = match (version, env::consts::OS, env::consts::ARCH) {
            ("8.12.0", "linux", "x86_64") => (
                "gitleaks-8.12.0-linux-x86_64",
                "9ed4271ffbfa04feec1423eb56154ad22c11ac4cae698f0115c1a064d4553524",
            ),
            // Add more supported versions here
            _ => ("gitleaks-unknown-unknown-unknown", "UNSUPPORTED"),
        };

        GitleaksConfig {
            filename: filename.to_string(),
            download_url: format!(
                "https://raw.githubusercontent.com/leaktk/bin/main/bin/{}",
                filename
            ),
            checksum: checksum.to_string(),
            version: version.to_string(),
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
