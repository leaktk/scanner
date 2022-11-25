pub mod config;
pub mod errors;
pub mod listner;
pub mod scanner;

use crate::config::Config;
use crate::errors::Error;
use crate::listner::Listner;
use crate::scanner::Scanner;
use clap::Parser;

#[derive(Parser)]
struct Args {
    #[arg(short, long)]
    config: String,
}

fn main() -> Result<(), Error> {
    let config = Config::load(&Args::parse().config)?;
    let mut scanner = Scanner::new(&config.scanner);

    for request in Listner::new() {
        let result = scanner.scan(&request);
        println!("{}", serde_json::to_string(&result).unwrap());
    }

    Ok(())
}
