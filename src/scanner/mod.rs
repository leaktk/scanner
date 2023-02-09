pub mod patterns;
pub mod proto;
pub mod providers;

mod gitleaks;

use crate::config::ScannerConfig;
use gitleaks::Gitleaks;
use log::{debug, error, info, warn};
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
            fs::remove_dir_all(scan_dir).expect("Could not remove scan dir!");
        }
    }

    fn start_scan(&self, req: &Request, scan_dir: &Path) -> Response {
        Response {
            id: Uuid::new_v4().to_string(),
            request: ResponseRequest {
                id: req.id.to_string(),
            },
            error: None,
            results: self
                .gitleaks
                .git_scan(scan_dir, &req.options)
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
        }
    }

    fn error_response(&self, req: &Request, error: &str) -> Response {
        Response {
            id: Uuid::new_v4().to_string(),
            request: ResponseRequest {
                id: req.id.to_string(),
            },
            results: Vec::new(),
            error: Some(error.to_string()),
        }
    }

    pub fn scan(&self, req: &Request) -> Response {
        self.refresh_stale_patterns();

        let scan_dir = self.scan_dir(&req);
        let clone_result = self.providers.clone(&req, &scan_dir);

        let resp = if clone_result.ok {
            debug!("Ok clone result msg: {}", clone_result.msg);
            self.start_scan(&req, &scan_dir)
        } else {
            self.error_response(&req, &clone_result.msg)
        };

        self.clean_up(&req, &scan_dir);

        resp
    }
}
