use std::env;
use std::fs;
use std::path::Path;
use std::path::PathBuf;

use serde::Deserialize;
use thiserror::Error;

#[derive(Error, Debug)]
pub enum ConfigError {
    #[error("could not read config file")]
    CouldNotRead(#[from] std::io::Error),

    #[error("invalid config")]
    InvalidConfig(#[from] toml::de::Error),
}

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
            ("8.12.0", "darwin", "arm64") => (
                "gitleaks-8.12.0-darwin-arm64",
                "da4e64fe24d2ed41e2472f67acdedbf6adadf8e6c7620ce037dc61d7b85859a7",
            ),
            ("8.12.0", "darwin", "x86_64") => (
                "gitleaks-8.12.0-darwin-x86_64",
                "708de7052fb76a2e61273d0b6210e717a3fac85e955ec1163d17e6bfe864fbfd",
            ),
            ("8.12.0", "linux", "x86_64") => (
                "gitleaks-8.12.0-linux-x86_64",
                "9ed4271ffbfa04feec1423eb56154ad22c11ac4cae698f0115c1a064d4553524",
            ),
            ("8.12.0", "windows", "x86_64") => (
                "gitleaks-8.12.0-windows-x86_64.exe",
                "d7a49162a15133958d3e4b8c0967581fe3069e81847978ad5b5f6903a9f6fa88",
            ),
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
        // Use user's cache_dir which is generally longer lived than the tmp_dir
        // or fall back on tmpdir if that doesn't exist
        if let Some(cache_dir) = dirs::cache_dir() {
            cache_dir.join("leaktk")
        } else {
            env::temp_dir().join("leaktk")
        }
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
    /// Load the config from a `&str`
    pub fn from_str(raw: &str) -> Result<Config, ConfigError> {
        Ok(toml::from_str(raw)?)
    }

    /// Load the config from a file path
    pub fn load_file(path: &Path) -> Result<Config, ConfigError> {
        Config::from_str(&fs::read_to_string(path)?)
    }

    /// Load the config from a provided file path or fall back on defaults
    pub fn load(path: Option<String>) -> Result<Config, ConfigError> {
        if let Some(path) = path {
            return Config::load_file(&Path::new(&path));
        }

        if let Some(config_dir) = dirs::config_dir() {
            let path = config_dir.join("leaktk").join("config.toml");

            if path.exists() {
                return Config::load_file(&path);
            }
        }

        #[cfg(any(target_os = "linux", target_os = "macos"))]
        {
            let path = Path::new("/etc/leaktk/config.toml");
            if path.exists() {
                return Config::load_file(&path);
            }
        }

        Ok(Default::default())
    }
}
