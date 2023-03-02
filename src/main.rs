pub mod config;
pub mod errors;
pub mod listener;
pub mod logger;
pub mod parser;
pub mod scanner;

use crate::config::Config;
use crate::errors::Error;
use crate::listener::Listener;
use crate::logger::Logger;
use crate::scanner::patterns::Patterns;
use crate::scanner::providers::Providers;
use crate::scanner::Scanner;

fn main() -> Result<(), Error> {
    let config = Config::load(parser::args().config)?;

    Logger::init(&config.logger)?;

    let patterns = Patterns::new(&config.scanner);
    let providers = Providers::new();
    let scanner = Scanner::new(&config.scanner, &providers, &patterns);

    for request in Listener::new() {
        let result = scanner.scan(&request);
        println!("{}", serde_json::to_string(&result).unwrap());
    }

    Ok(())
}
