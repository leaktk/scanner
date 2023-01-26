pub mod patterns;
pub mod proto;
pub mod providers;

mod gitleaks;

use crate::config::ScannerConfig;
use gitleaks::Gitleaks;
use log::{error, info, warn};
use patterns::Patterns;
use proto::{
    GitCommit, GitCommitAuthor, Lines, Request, Response, ResponseRequest, Result as ScanResult,
    Rule, Source,
};
use providers::Providers;
use std::fs;
use std::path::{Path, PathBuf};
use uuid::Uuid;

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
        scanner.refresh_stale_patterns();
        scanner
    }

    fn refresh_stale_patterns(&self) {
        if let Err(err) = self.patterns.refresh_if_stale() {
            error!("{}", err);
            if !self.patterns.exists() {
                // There's really nothing we can do here. If there aren't any
                // patterns then this error isn't recoverable at this point
                // in time.
                panic!("Could not find patterns!");
            } else {
                warn!("Falling back on stale patterns!");
            }
        }
    }

    fn reset_scans_dir(&self) {
        info!("Resetting scan dir");

        if self.scans_dir().as_path().exists() {
            fs::remove_dir_all(self.scans_dir())
                // If the scan dir can't be removed. The code shouldn't run
                .expect("Could not clear scans dir!");
        }
    }

    // The dir for scan folders
    fn scans_dir(&self) -> PathBuf {
        self.config.workdir.join("scans")
    }

    // The dir for a specific scan folder
    fn scan_dir(&self) -> PathBuf {
        self.scans_dir().join(Uuid::new_v4().to_string())
    }

    fn start_git_scan(&self, id: &str, url: &str, scan_dir: &Path) -> Response {
        let gitleaks_results = self.gitleaks.git_scan(scan_dir);

        Response {
            id: Uuid::new_v4().to_string(),
            request: ResponseRequest { id: id.to_string() },
            error: None,
            results: gitleaks_results
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
                        url: url.to_string(),
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
        }
    }

    fn error_response(&self, id: &str, error: &str) -> Response {
        Response {
            id: Uuid::new_v4().to_string(),
            request: ResponseRequest { id: id.to_string() },
            results: Vec::new(),
            error: Some(error.to_string()),
        }
    }

    pub fn scan(&self, req: &Request) -> Response {
        self.refresh_stale_patterns();
        let scan_dir = self.scan_dir();

        let resp = match req {
            Request::Git { id, url, .. } => match self.providers.clone(&req, &scan_dir) {
                Err(err) => self.error_response(&id, &err.to_string()),
                Ok(output) => {
                    if output.status.success() {
                        self.start_git_scan(&id, &url, &scan_dir)
                    } else {
                        let error = String::from_utf8_lossy(&output.stderr);
                        self.error_response(&id, &error)
                    }
                }
            },
        };

        if scan_dir.exists() {
            fs::remove_dir_all(scan_dir).expect("Could not remove scan dir!");
        }

        resp
    }
}
