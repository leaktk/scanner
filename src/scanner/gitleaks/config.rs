use std::fs;
use std::path::Path;

use regex::Regex;
use serde::{self, Deserialize, Serialize};
use thiserror::Error;

fn filter_regex(values: &Vec<String>) -> Vec<String> {
    values
        .iter()
        .filter(|value| Regex::new(value).is_ok())
        .map(String::to_string)
        .collect()
}

#[derive(Error, Debug)]
pub enum ConfigError {
    #[error("could not read config file")]
    CouldNotRead(#[from] std::io::Error),

    #[error("invalid config: {0}")]
    InvalidConfig(#[from] toml::de::Error),
}

#[derive(Debug, Deserialize, Serialize)]
struct Extends {
    path: String,
}

#[derive(Debug, Deserialize, Serialize)]
struct Allowlist {
    description: Option<String>,
    regexes: Option<Vec<String>>,
    #[serde(rename = "regexTarget")]
    regex_target: Option<String>,
    paths: Option<Vec<String>>,
    commits: Option<Vec<String>>,
    stopwords: Option<Vec<String>>,
}

impl Allowlist {
    pub fn clean_regex_fields(&mut self) {
        // Make sure that the regexes are valid regexes
        if let Some(regexes) = &self.regexes {
            self.regexes = Some(filter_regex(regexes));
        }

        // Make sure that the paths are valid regexes
        if let Some(paths) = &self.paths {
            self.paths = Some(filter_regex(paths));
        }
    }
}

// This determines which fields we allow from a .gitleaks.toml in a repo
#[derive(Debug, Deserialize)]
struct RestrictedConfig {
    allowlist: Option<Allowlist>,
}

impl RestrictedConfig {
    fn from_str(raw: &str) -> Result<Self, ConfigError> {
        let mut config: Self = toml::from_str(raw)?;

        if let Some(ref mut allowlist) = config.allowlist {
            allowlist.clean_regex_fields();
        }

        Ok(config)
    }

    pub fn load_file(path: &Path) -> Result<Self, ConfigError> {
        Self::from_str(&fs::read_to_string(path)?)
    }
}

/// # Gitleaks Repo Config
///
/// A repo may have a .gitleaks.toml in it with custom rules. Given that this
/// scanner and tooling relies on specific tags to work properly, we use a
/// custom config file, but we don't want to ignore things allowed in the
/// .gitleaks.toml.
///
/// This object creates a small gitleaks.toml that's saved under
/// `{workspace.config_dir}/gitleaks.toml` that looks something like:
///
/// ```toml
/// [extends]
/// path = "/path/to/global/config/from/the/pattern/server.toml"
///
/// [allowlist]
/// # items from the repo's gitleaks.toml
/// ```
///
/// This way people can use either their .gitleaks.toml or a .gitleaksignore
/// file for ignoring things in their repo.

#[derive(Debug, Serialize)]
pub struct GitleaksRepoConfig {
    extends: Extends,
    allowlist: Option<Allowlist>,
}

impl GitleaksRepoConfig {
    pub fn new(
        global_config_path: &Path,
        repo_gitleaks_toml_path: &Path,
    ) -> Result<Self, ConfigError> {
        let restricted_config = RestrictedConfig::load_file(repo_gitleaks_toml_path)?;
        let repo_config = Self {
            allowlist: restricted_config.allowlist,
            extends: Extends {
                path: global_config_path.display().to_string(),
            },
        };

        Ok(repo_config)
    }

    pub fn is_empty(&self) -> bool {
        self.allowlist.is_none()
    }
}
