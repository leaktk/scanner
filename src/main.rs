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
use crate::scanner::Scanner;

fn main() -> Result<(), Error> {
    let config = match parser::args().config {
        Some(config_path) => Config::load(&config_path)?,
        None => Default::default(),
    };

    Logger::init(&config.logger)?;

    let scanner = Scanner::new(&config.scanner);

    for request in Listner::new() {
        let result = scanner.scan(&request);
        println!("{}", serde_json::to_string(&result).unwrap());
    }

    Ok(())
}
