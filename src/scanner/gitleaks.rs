use super::patterns::Patterns;
use super::proto::GitLeaksResult;
use crate::config::ScannerConfig;
use std::fs::{self, File};
use std::io::Write;
use std::os::unix::fs::PermissionsExt;
use std::path::{Path, PathBuf};
use std::process::Command;

use log::info;
use ring::digest::{Context, SHA256};

fn gitleaks_path(config: &ScannerConfig) -> PathBuf {
    // TODO: better error handling for this
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

pub fn scan(config: &ScannerConfig, patterns: &Patterns, scan_dir: &Path) -> Vec<GitLeaksResult> {
    let results = Command::new(gitleaks_path(config))
        .arg("detect")
        .arg("--report-path=/dev/stdout")
        .arg("--report-format=json")
        .arg("--config")
        .arg(&patterns.path)
        .arg("--source")
        .arg(scan_dir)
        .output()
        .expect("Could not run scan");

    // TODO: Better error handling
    let raw_results = String::from_utf8(results.stdout).unwrap();
    serde_json::from_str(&raw_results).unwrap()
}
