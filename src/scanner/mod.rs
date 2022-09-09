pub mod proto;

mod patterns;
mod providers;

use crate::config::ScannerConfig;
use proto::{Kind, Request, Response};
use std::path::{Path, PathBuf};

pub struct Scanner<'s> {
    config: &'s ScannerConfig,
    scan_count: u32,
}

impl<'s> Scanner<'s> {
    pub fn new(config: &'s ScannerConfig) -> Scanner<'s> {
        let scanner = Scanner {
            config: config,
            scan_count: 0,
        };

        // TODO: have this respect the interval and be called in a background
        // thread after initial setup
        patterns::refresh(&scanner.config);

        // TODO: clear out the scans dir in the workdir on initial setup
        //
        scanner
    }

    // Provide a unique path to place files in for scanning
    fn scan_dir(&mut self) -> PathBuf {
        self.scan_count += 1;

        self.config
            .workdir
            .join("scans")
            .join(self.scan_count.to_string())
            .join("files")
    }

    fn fetch_files(&mut self, req: &Request, scan_dir: &Path) {
        match req.kind {
            // TODO: figure out how to turn this into a trait
            Kind::Git => providers::git::clone(req.url, scan_dir),
        }
    }

    pub fn scan<'r>(&mut self, req: &'r Request) -> Vec<Response<'r>> {
        let scan_dir = self.scan_dir();

        // TODO: hand these off to a worker pool
        self.fetch_files(&req, scan_dir.as_path());
        // self.start_scan(&req, scan_dir.as_path())

        // TODO: figure out how to stream responses from the worker pool
        vec![Response::new(req)]
    }
}
