use super::patterns::Patterns;
use super::proto::GitLeaksResult;
use super::providers::Providers;
use crate::config::ScannerConfig;
use log::info;
use ring::digest::{Context, SHA256};
use std::fs::{self, File};
use std::io::Write;
use std::os::unix::fs::PermissionsExt;
use std::path::{Path, PathBuf};
use std::process::Command;

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
            config: config,
            providers: providers,
            patterns: patterns,
        }
    }

    #[inline]
    fn download_gitleaks(&self, bindir: &Path, binpath: &Path) {
        fs::create_dir_all(&bindir).expect("Could not create bin file directory!");

        let req = reqwest::blocking::get(&self.config.gitleaks.download_url).unwrap();
        let data = req.bytes().unwrap();
        let mut bin = File::create(bindir.join(&self.config.gitleaks.filename)).unwrap();

        bin.write_all(&data).unwrap();

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
            fs::remove_file(binpath).unwrap();
            panic!("Invalid gitleaks digest!");
        }

        let mut perms = fs::metadata(&binpath).unwrap().permissions();
        perms.set_mode(0o770);
        fs::set_permissions(&binpath, perms).unwrap();

        info!("{} downloaded!", &self.config.gitleaks.filename);
    }

    fn gitleaks_path(&self) -> PathBuf {
        let bindir = self.config.workdir.join("bin");
        let binpath = bindir.join(&self.config.gitleaks.filename);

        if !binpath.exists() {
            self.download_gitleaks(&bindir, &binpath);
        }

        binpath
    }

    fn gitleaks_log_opts(&self, scan_dir: &Path) -> Vec<String> {
        let shallow_commits = self.providers.git.shallow_commits(&scan_dir);

        if shallow_commits.len() > 0 {
            let exclude_commits: Vec<String> =
                shallow_commits.iter().map(|s| format!("^{s}")).collect();

            let log_opts = format!(
                "--full-history --simplify-merges --show-pulls --all {}",
                exclude_commits.join(" ")
            );
            vec!["--log-opts".to_string(), log_opts]
        } else {
            vec![]
        }
    }

    pub fn git_scan(&self, scan_dir: &Path) -> Vec<GitLeaksResult> {
        let gitleaks_path = self.gitleaks_path();
        let mut args = vec![
            "detect".to_string(),
            "--report-path=/dev/stdout".to_string(),
            "--report-format=json".to_string(),
            "--config".to_string(),
            self.patterns.gitleaks_patterns_path.display().to_string(),
            "--source".to_string(),
            scan_dir.display().to_string(),
        ];

        args.extend(self.gitleaks_log_opts(&scan_dir));

        info!("Running: {} {}", gitleaks_path.display(), args.join(" "));
        let results = Command::new(&gitleaks_path)
            .args(args)
            .output()
            .expect("Could not run scan");

        let raw_results = String::from_utf8_lossy(&results.stdout);
        serde_json::from_str(&raw_results).unwrap()
    }
}
