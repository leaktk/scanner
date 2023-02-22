pub mod config;
pub mod errors;
pub mod listner;
pub mod logging;
pub mod parser;
pub mod scanner;

use crate::config::Config;
use crate::errors::Error;
use crate::listner::Listner;
use crate::logging::Logger;
use crate::scanner::patterns::Patterns;
use crate::scanner::providers::Providers;
use crate::scanner::Scanner;

fn main() -> Result<(), Error> {
    let config = match parser::args().config {
        Some(config_path) => Config::load(&config_path)?,
        None => Config::default_load()?,
    };

    Logger::init(&config.logger)?;

    let patterns = Patterns::new(&config.scanner);
    let providers = Providers::new();
    let scanner = Scanner::new(&config.scanner, &providers, &patterns);

    for request in Listner::new() {
        let result = scanner.scan(&request);
        println!("{}", serde_json::to_string(&result).unwrap());
    }

    Ok(())
}
