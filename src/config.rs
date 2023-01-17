use crate::errors::Error;
use serde::Deserialize;
use std::env;
use std::fs;
use std::path::PathBuf;

#[derive(Debug, Deserialize)]
pub struct GitleaksConfig {
    pub filename: String,
    pub version: String,
    pub download_url: String,
    pub checksum: String,
}

// Generally you shouldn't override the gitleaks config but this provides some flexibility and
// provides a consistent way to access the details in an OS agnoistic way. Just make sure it's
// compatible with the version defined in the source.
//
// It also doesn't make a whole lot of sense to set the default for one value
// and not change the others, hence no field level defaults.
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

#[derive(Debug, Deserialize)]
pub struct PatternServerConfig {
    #[serde(default = "PatternServerConfig::default_url")]
    pub url: String,
}

impl PatternServerConfig {
    fn default_url() -> String {
        "https://raw.githubusercontent.com/leaktk/patterns/main/target".to_string()
    }
}

impl Default for PatternServerConfig {
    fn default() -> Self {
        PatternServerConfig {
            url: PatternServerConfig::default_url(),
        }
    }
}

#[derive(Debug, Deserialize)]
pub struct PatternsConfig {
    #[serde(default)]
    pub server: PatternServerConfig,
    #[serde(default = "PatternsConfig::default_refresh_interval")]
    pub refresh_interval: u64,
}

impl PatternsConfig {
    fn default_refresh_interval() -> u64 {
        43200
    }
}

impl Default for PatternsConfig {
    fn default() -> Self {
        PatternsConfig {
            server: Default::default(),
            refresh_interval: PatternsConfig::default_refresh_interval(),
        }
    }
}

#[derive(Debug, Deserialize)]
pub struct LoggerConfig {
    #[serde(default = "LoggerConfig::default_level")]
    pub level: log::Level,
}

impl LoggerConfig {
    fn default_level() -> log::Level {
        log::Level::Info
    }
}

impl Default for LoggerConfig {
    fn default() -> Self {
        LoggerConfig {
            level: LoggerConfig::default_level(),
        }
    }
}

#[derive(Debug, Deserialize)]
pub struct ScannerConfig {
    #[serde(default)]
    pub gitleaks: GitleaksConfig,
    #[serde(default = "ScannerConfig::default_workdir")]
    pub workdir: PathBuf,
    #[serde(default)]
    pub patterns: PatternsConfig,
}

impl ScannerConfig {
    fn default_workdir() -> PathBuf {
        env::temp_dir().join("leaktk")
    }
}

impl Default for ScannerConfig {
    fn default() -> Self {
        ScannerConfig {
            gitleaks: Default::default(),
            workdir: ScannerConfig::default_workdir(),
            patterns: Default::default(),
        }
    }
}

#[derive(Debug, Default, Deserialize)]
pub struct Config {
    #[serde(default)]
    pub logger: LoggerConfig,
    #[serde(default)]
    pub scanner: ScannerConfig,
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
