pub mod patterns;
pub mod proto;
pub mod providers;

mod gitleaks;

use std::fs;
use std::path::{Path, PathBuf};

use log::{debug, error, info, warn};
use thiserror::Error;
use uuid::Uuid;

use crate::config::ScannerConfig;

use gitleaks::{Gitleaks, GitleaksError};
use patterns::{Patterns, PatternsError};
use proto::{
    GitCommit, GitCommitAuthor, Lines, Request, Response, ResponseRequest, Result as ScanResult,
    Rule, Source,
};
use providers::{ProviderError, Providers};

#[derive(Error, Debug)]
pub enum ScannerError {
    #[error(transparent)]
    CouldNotClone(#[from] ProviderError),

    #[error(transparent)]
    CouldNotFetchPatterns(#[from] PatternsError),

    #[error(transparent)]
    GitleaksFailed(#[from] GitleaksError),
}

pub struct Scanner<'s> {
    config: &'s ScannerConfig,
    patterns: &'s Patterns,
    providers: &'s Providers,
    gitleaks: Gitleaks<'s>,
}

impl<'s> Scanner<'s> {
    pub fn new(
        config: &'s ScannerConfig,
        providers: &'s Providers,
        patterns: &'s Patterns,
    ) -> Scanner<'s> {
        let scanner = Scanner {
            config: config,
            patterns: patterns,
            providers: providers,
            gitleaks: Gitleaks::new(&config, &providers, &patterns),
        };

        scanner.reset_scans_dir();
        scanner
    }

    fn reset_scans_dir(&self) {
        info!("resetting scan dir");

        if self.scans_dir().as_path().exists() {
            fs::remove_dir_all(self.scans_dir()).unwrap_or_else(|err| {
                error!("could not reset scan dir: {}", err);
            });
        }
    }

    // The dir for scan folders
    fn scans_dir(&self) -> PathBuf {
        self.config.workdir.join("scans")
    }

    // The dir for a specific scan folder
    fn scan_dir(&self, req: &Request) -> PathBuf {
        if req.is_local() {
            PathBuf::from(&req.target)
        } else {
            self.scans_dir().join(Uuid::new_v4().to_string())
        }
    }

    // Clean up after a scan is over
    fn clean_up(&self, req: &Request, scan_dir: &PathBuf) {
        if !req.is_local() && scan_dir.exists() {
            fs::remove_dir_all(scan_dir).unwrap_or_else(|err| {
                error!("could not remove {}: {}", scan_dir.display(), err);
            });
        }
    }

    fn start_scan(&self, req: &Request, scan_dir: &Path) -> Result<Response, ScannerError> {
        Ok(Response {
            id: Uuid::new_v4().to_string(),
            request: ResponseRequest {
                id: req.id.to_string(),
            },
            error: None,
            results: self
                .gitleaks
                .git_scan(scan_dir, &req.options)?
                .iter()
                .map(|r| ScanResult {
                    context: r.context.clone(),
                    target: r.target.clone(),
                    entropy: r.entropy,
                    rule: Rule {
                        id: r.rule_id.clone(),
                        description: r.rule_description.clone(),
                        tags: r.rule_tags.clone(),
                    },
                    source: Source::Git {
                        target: req.target.to_string(),
                        path: r.source_path.clone(),
                        lines: Lines {
                            start: r.source_lines_start.clone(),
                            end: r.source_lines_end.clone(),
                        },
                        commit: GitCommit {
                            id: r.source_commit_id.clone(),
                            date: r.source_commit_date.clone(),
                            message: r.source_commit_message.clone(),
                            author: GitCommitAuthor {
                                name: r.source_commit_author_name.clone(),
                                email: r.source_commit_author_name.clone(),
                            },
                        },
                    },
                })
                .collect(),
        })
    }

    fn error_response(&self, req: &Request, err: ScannerError) -> Response {
        error!("{}", err);

        Response {
            id: Uuid::new_v4().to_string(),
            request: ResponseRequest {
                id: req.id.to_string(),
            },
            results: Vec::new(),
            error: Some(err.to_string()),
        }
    }

    pub fn scan(&self, req: &Request) -> Response {
        if let Err(err) = self.patterns.refresh_if_stale() {
            if self.patterns.exists() {
                warn!("using on stale patterns");
            }

            return self.error_response(&req, ScannerError::from(err));
        }

        let scan_dir = self.scan_dir(&req);

        let resp = self
            .providers
            .clone(&req, &scan_dir)
            .map_err(ScannerError::CouldNotClone)
            .and_then(|msg| {
                debug!("clone success: {}", msg);
                self.start_scan(&req, &scan_dir)
            })
            .unwrap_or_else(|err| self.error_response(&req, err));

        self.clean_up(&req, &scan_dir);

        resp
    }
}
