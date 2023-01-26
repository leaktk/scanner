use super::patterns::Patterns;
use super::proto::GitLeaksResult;
use super::providers::git::Git;
use crate::config::ScannerConfig;
use log::info;
use ring::digest::{Context, SHA256};
use std::fs::{self, File};
use std::io::Write;
use std::os::unix::fs::PermissionsExt;
use std::path::{Path, PathBuf};
use std::process::Command;

pub struct Gitleaks;

impl Gitleaks {
    pub fn new() -> Gitleaks {
        Gitleaks {}
    }

    fn gitleaks_path(&self, config: &ScannerConfig) -> PathBuf {
        let bindir = config.workdir.join("bin");
        let binpath = bindir.join(&config.gitleaks.filename);

        if binpath.exists() {
            return binpath;
        }

        fs::create_dir_all(&bindir).expect("Could not create bin file directory!");

        let req = reqwest::blocking::get(&config.gitleaks.download_url).unwrap();
        let data = req.bytes().unwrap();
        let mut bin = File::create(bindir.join(&config.gitleaks.filename)).unwrap();

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

        if hex_digest != config.gitleaks.checksum {
            fs::remove_file(binpath).unwrap();
            panic!("Invalid gitleaks digest!");
        }

        let mut perms = fs::metadata(&binpath).unwrap().permissions();
        perms.set_mode(0o770);
        fs::set_permissions(&binpath, perms).unwrap();

        info!("{} downloaded!", &config.gitleaks.filename);

        return binpath;
    }

    fn gitleaks_log_opts(&self, git: &Git, scan_dir: &Path) -> Vec<String> {
        let shallow_commits = git.shallow_commits(&scan_dir);

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

    pub fn git_scan(
        &self,
        config: &ScannerConfig,
        git: &Git,
        patterns: &Patterns,
        scan_dir: &Path,
    ) -> Vec<GitLeaksResult> {
        let results = Command::new(self.gitleaks_path(config))
            .arg("detect")
            .arg("--report-path=/dev/stdout")
            .arg("--report-format=json")
            .arg("--config")
            .arg(&patterns.path)
            .arg("--source")
            .arg(scan_dir)
            .args(self.gitleaks_log_opts(&git, &scan_dir))
            .output()
            .expect("Could not run scan");

        let raw_results = String::from_utf8(results.stdout).unwrap();
        serde_json::from_str(&raw_results).unwrap()
    }
}
