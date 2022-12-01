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
use clap::Parser;

#[derive(Parser)]
struct Args {
    #[arg(short, long)]
    config: String,
}

fn main() -> Result<(), Error> {
    let config = Config::load(&Args::parse().config)?;

    Logger::init(&config.logger)?;

    let mut scanner = Scanner::new(&config.scanner);

    for request in Listner::new() {
        let result = scanner.scan(&request);
        println!("{}", serde_json::to_string(&result).unwrap());
    }

    Ok(())
}
