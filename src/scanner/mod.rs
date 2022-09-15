pub mod proto;

mod patterns;
mod providers;

use std::fs;
use std::path::{Path, PathBuf};
use std::time::{Instant,Duration};
use crate::config::ScannerConfig;
use proto::{Kind, Request, Response};

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
        fs::remove_dir_all(self.scans_dir())
            // If the scan dir can't be removed. The code shouldn't run
            .expect("Could not clear scans dir!");
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

    fn fetch_files(&self, req: &Request, files_dir: &Path) {
        // TODO: audit log statment
        match req.kind {
            // TODO: figure out how to turn this into a trait
            Kind::GitRepoURL => providers::git::clone(&req, files_dir),
        }
    }

    // TODO: Might implement this as a GitLeaks trait
    fn start_scan<'r>(&self, req: &'r Request, files_dir: &Path) -> Vec<Response<'r>> {
        // TODO: audit log statment
        println!("TODO: start gitleaks scan on {}!", files_dir.display());
        // gitleaks::scan(self.config, req, files_dir) -> vec![Response::new(req)]
        vec![Response::new(req)]
    }

    pub fn scan<'r>(&mut self, req: &'r Request) -> Vec<Response<'r>> {
        self.scan_count += 1;
        let files_dir = self.scan_job_files_dir(self.scan_count);

        // NOTE: if threading this, scans should not run while a refresh is
        self.refresh_stale_patterns();

        self.fetch_files(&req, files_dir.as_path());
        self.start_scan(&req, files_dir.as_path())
    }
}
