pub mod proto;

mod gitleaks;
mod patterns;
mod providers;

use crate::config::ScannerConfig;
use proto::{
    GitCommit, GitCommitAuthor, GitOptions, Lines, Request, Response, ResponseRequest,
    Result as ScanResult, Rule, Source,
};
use std::fs;
use std::path::{Path, PathBuf};
use std::time::{Duration, Instant};

pub struct Scanner<'s> {
    config: &'s ScannerConfig,
    scan_count: u32,
    last_patterns_refresh: Option<Instant>,
    refresh_interval: Duration,
}

impl<'s> Scanner<'s> {
    pub fn new(config: &'s ScannerConfig) -> Scanner<'s> {
        let mut scanner = Scanner {
            config: config,
            scan_count: 0,
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
        // TODO: Audit log statement
        if Path::exists(&self.scans_dir()) {
            fs::remove_dir_all(self.scans_dir())
                // If the scan dir can't be removed. The code shouldn't run
                .expect("Could not clear scans dir!");
        }
    }

    fn scans_dir(&self) -> PathBuf {
        self.config.workdir.join("scans")
    }

    fn scan_job_dir(&self, scan_id: u32) -> PathBuf {
        self.scans_dir().join(scan_id.to_string())
    }

    fn scan_job_files_dir(&self, scan_id: u32) -> PathBuf {
        self.scan_job_dir(scan_id).join("files")
    }

    fn start_git_scan(
        &self,
        id: &String,
        url: &String,
        options: &Option<GitOptions>,
        files_dir: &Path,
    ) -> Response {
        let gitleaks_results = gitleaks::scan(&self.config, files_dir, options);

        Response {
            id: "TODO".to_string(),
            request: ResponseRequest { id: id.clone() },
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
                        url: url.clone(),
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

    pub fn scan(&mut self, req: &Request) -> Response {
        self.scan_count += 1;
        let files_dir = self.scan_job_files_dir(self.scan_count);

        self.refresh_stale_patterns();

        match req {
            Request::Git { id, url, options } => {
                providers::git::clone(&id, &url, &options, files_dir.as_path());
                self.start_git_scan(&id, &url, &options, files_dir.as_path())
            }
        }
    }
}
