pub mod proto;

mod gitleaks;
mod patterns;
mod providers;

use crate::config::ScannerConfig;
use log::info;
use proto::{
    GitCommit, GitCommitAuthor, Lines, Request, Response, ResponseRequest, Result as ScanResult,
    Rule, Source,
};
use std::fs;
use std::path::{Path, PathBuf};
use std::time::{Duration, Instant};
use uuid::Uuid;

pub struct Scanner<'s> {
    config: &'s ScannerConfig,
    last_patterns_refresh: Option<Instant>,
    refresh_interval: Duration,
}

impl<'s> Scanner<'s> {
    pub fn new(config: &'s ScannerConfig) -> Scanner<'s> {
        let mut scanner = Scanner {
            config: config,
            // TODO: look this up from the file timestamp and make this not optional
            last_patterns_refresh: None,
            refresh_interval: Duration::from_secs(config.patterns.refresh_interval),
        };

        scanner.reset_scans_dir();
        scanner.refresh_stale_patterns();
        scanner
    }

    fn refresh_stale_patterns(&mut self) {
        if match self.last_patterns_refresh {
            None => true,
            Some(last) => last.duration_since(Instant::now()) > self.refresh_interval,
        } {
            patterns::refresh(&self.config);
            self.last_patterns_refresh = Some(Instant::now());
        }
    }

    fn reset_scans_dir(&self) {
        info!("Resetting scan dir");
        if Path::exists(&self.scans_dir()) {
            fs::remove_dir_all(self.scans_dir())
                // If the scan dir can't be removed. The code shouldn't run
                .expect("Could not clear scans dir!");
        }
    }

    fn scans_dir(&self) -> PathBuf {
        self.config.workdir.join("scans")
    }

    fn scan_dir(&self) -> PathBuf {
        self.scans_dir().join(Uuid::new_v4().to_string())
    }

    fn start_git_scan(&self, id: &str, url: &str, scan_dir: &Path) -> Response {
        let gitleaks_results = gitleaks::scan(&self.config, scan_dir);

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

    pub fn scan(&mut self, req: &Request) -> Response {
        self.refresh_stale_patterns();

        match req {
            Request::Git { id, url, options } => {
                let scan_dir = self.scan_dir();
                let result = providers::git::clone(&url, &options, scan_dir.as_path());

                match result {
                    Err(err) => self.error_response(&id, &err.to_string()),
                    Ok(output) => {
                        if output.status.success() {
                            self.start_git_scan(&id, &url, scan_dir.as_path())
                        } else {
                            let error = String::from_utf8_lossy(&output.stderr);
                            self.error_response(&id, &error)
                        }
                    }
                }
            }
        }
    }
}
