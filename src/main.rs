pub mod config;
pub mod errors;
pub mod listner;
pub mod logging;
pub mod scanner;

use crate::config::Config;
use crate::errors::Error;
use crate::listner::Listner;
use crate::logging::Logger;
use crate::scanner::Scanner;
use std::env;

fn main() -> Result<(), Error> {
    // TODO write a simple parser for this
    assert_eq!(&env::args().nth(1).expect("--config present"), "--config");
    let config = Config::load(&env::args().nth(2).expect("config path present"))?;

    Logger::init(&config.logger)?;

    let scanner = Scanner::new(&config.scanner);

    for request in Listner::new() {
        let result = scanner.scan(&request);
        println!("{}", serde_json::to_string(&result).unwrap());
    }

    Ok(())
}
