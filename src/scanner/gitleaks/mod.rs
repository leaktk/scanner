mod config;

use std::fs;
use std::fs::File;
#[cfg(target_family = "unix")]
use std::os::unix::fs::PermissionsExt;
use std::path::{Path, PathBuf};
use std::process::Command;

use log::{info, warn};
use ring::digest::{Context, SHA256};
use thiserror::Error;

use crate::config::ScannerConfig;

use super::patterns::Patterns;
use super::proto::{GitLeaksResult, RequestOptions};
use super::providers::Providers;
use super::workspace::Workspace;

use config::{ConfigError, GitleaksRepoConfig};

#[derive(Error, Debug)]
pub enum GitleaksError {
    #[error("could not set up gitleaks")]
    CouldNotSetUpGitleaks(#[from] std::io::Error),

    #[error("could not fetch gitleaks")]
    CouldNotFetchGitleaks(#[from] reqwest::Error),

    #[error("could not parse results")]
    CouldNotParseResults(#[from] serde_json::Error),

    #[error("could not complete scan")]
    CouldNotCompleteScan(#[source] std::io::Error),

    #[error("could not open results file")]
    CouldNotReadResultsFile(#[source] std::io::Error),

    #[error("invalid gitleaks digest")]
    InvalidGitleaksDigest,
}

#[derive(Error, Debug)]
pub enum RepoConfigError {
    #[error(transparent)]
    Invalid(#[from] ConfigError),

    #[error("could not set up repo config")]
    CouldNotSetUp(#[from] std::io::Error),

    #[error("could not serialize repo config")]
    CouldNotSerialize(#[from] toml::ser::Error),
}

pub struct Gitleaks<'g> {
    config: &'g ScannerConfig,
    providers: &'g Providers,
    patterns: &'g Patterns,
}

impl<'g> Gitleaks<'g> {
    pub fn new(
        config: &'g ScannerConfig,
        providers: &'g Providers,
        patterns: &'g Patterns,
    ) -> Gitleaks<'g> {
        Gitleaks {
            config,
            providers,
            patterns,
        }
    }

    #[inline]
    fn download_gitleaks(&self, bindir: &Path, binpath: &Path) -> Result<(), GitleaksError> {
        fs::create_dir_all(bindir)?;

        let data = reqwest::blocking::get(&self.config.gitleaks.download_url)?.bytes()?;
        fs::write(bindir.join(&self.config.gitleaks.filename), &data)?;

        let mut context = Context::new(&SHA256);
        context.update(&data);

        let hex_digest = context
            .finish()
            .as_ref()
            .iter()
            .map(|b| format!("{:02x}", b))
            .collect::<Vec<String>>()
            .join("");

        if hex_digest != self.config.gitleaks.checksum {
            fs::remove_file(binpath)?;
            return Err(GitleaksError::InvalidGitleaksDigest);
        }

        #[cfg(target_family = "unix")]
        {
            let mut perms = fs::metadata(binpath)?.permissions();
            perms.set_mode(0o770);
            fs::set_permissions(binpath, perms)?;
        }

        info!("{} downloaded", &self.config.gitleaks.filename);
        Ok(())
    }

    fn gitleaks_path(&self) -> Result<PathBuf, GitleaksError> {
        let bindir = self.config.workdir.join("bin");
        let binpath = bindir.join(&self.config.gitleaks.filename);

        if !binpath.exists() {
            self.download_gitleaks(&bindir, &binpath)?;
        }

        Ok(binpath)
    }

    fn gitleaks_log_opts(&self, scan_dir: &Path, options: &RequestOptions) -> Vec<String> {
        let mut log_opts = vec![
            "--full-history".to_string(),
            "--simplify-merges".to_string(),
            "--show-pulls".to_string(),
        ];

        if let Some(since) = &options.since {
            log_opts.push(format!("--since-as-filter={since}T00:00:00-00:00"));
        }

        if options.single_branch.unwrap_or(false) {
            // For now, depth is only supported for single branches until I
            // figure out a fast way to get n commits from each branch like
            // --depth does during a clone. I would rather risk over scanning
            // than underscanning
            if let Some(depth) = options.depth {
                log_opts.push(format!("--max-count={depth}"));
            }
        } else {
            log_opts.push("--all".to_string());
        }

        if let Some(branch) = &options.branch {
            log_opts.push(branch.to_string());
        }

        // For now when dealing with a local repo, don't exclude shallow
        // commits. The intent of excluding shallow commits was to avoid
        // over scanning during clones done by the scanner because it
        // trys to compensate by cloning a little deeper than requested.
        if !options.local.unwrap_or(false) {
            let to_exclude: Vec<String> = self
                .providers
                .git
                .shallow_commits(scan_dir)
                .iter()
                .map(|s| format!("^{s}"))
                .collect();

            log_opts.extend(to_exclude);
        }

        vec!["--log-opts".to_string(), log_opts.join(" ")]
    }

    fn setup_repo_config(&self, workspace: &Workspace) -> Result<PathBuf, RepoConfigError> {
        let global_config_path = &self.patterns.gitleaks_patterns_path;
        let repo_gitleaks_toml_path = workspace.scan_dir.join(".gitleaks.toml");

        if !repo_gitleaks_toml_path.exists() {
            return Ok(global_config_path.clone());
        }

        let repo_config = GitleaksRepoConfig::new(global_config_path, &repo_gitleaks_toml_path)?;

        // No reason to use it if it's empty
        if repo_config.is_empty() {
            return Ok(global_config_path.clone());
        }

        let repo_config_path = workspace.config_dir.join("gitleaks.toml");

        fs::create_dir_all(&workspace.config_dir)?;
        fs::create_dir_all(&workspace.results_dir)?;
        fs::write(&repo_config_path, toml::to_string(&repo_config)?)?;

        Ok(repo_config_path)
    }

    fn patterns_path(&self, workspace: &Workspace) -> PathBuf {
        match self.setup_repo_config(workspace) {
            Ok(repo_config_path) => repo_config_path,
            Err(err) => {
                warn!(
                    "could not configure custom repo config for {}: {}",
                    workspace, err,
                );

                self.patterns.gitleaks_patterns_path.clone()
            }
        }
    }

    pub fn git_scan(
        &self,
        workspace: &Workspace,
        options: &RequestOptions,
    ) -> Result<Vec<GitLeaksResult>, GitleaksError> {
        let gitleaks_path = self.gitleaks_path()?;
        let staged = options.staged.unwrap_or(false);
        let uncommitted = options.uncommitted.unwrap_or(staged);
        let report_path = workspace.results_dir.join("gitleaks-results.json");

        let mut args = vec![
            "--report-path".to_string(),
            report_path.display().to_string(),
            "--report-format=json".to_string(),
            "--config".to_string(),
            self.patterns_path(workspace).display().to_string(),
            "--source".to_string(),
            workspace.scan_dir.display().to_string(),
        ];

        if uncommitted {
            args.push("protect".to_string());

            if staged {
                args.push("--staged".to_string());
            }
        } else {
            args.push("detect".to_string());
            args.extend(self.gitleaks_log_opts(&workspace.scan_dir, options));
        }

        info!("running {} '{}'", gitleaks_path.display(), args.join("' '"));
        Command::new(&gitleaks_path)
            .args(args)
            .output()
            .map_err(GitleaksError::CouldNotCompleteScan)?;

        let results = File::open(report_path).map_err(GitleaksError::CouldNotReadResultsFile)?;

        Ok(serde_json::from_reader(results)?)
    }
}
