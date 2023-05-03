pub mod config;
pub mod listener;
pub mod logger;
pub mod parser;
pub mod scanner;

use std::panic;
use log::error;

use crate::config::Config;
use crate::listener::Listener;
use crate::logger::Logger;
use crate::scanner::patterns::Patterns;
use crate::scanner::providers::Providers;
use crate::scanner::Scanner;

use anyhow::Result;

fn main() -> Result<()> {
    let config = Config::load(parser::args().config)?;

    Logger::init(&config.logger)?;

    // Use the logger for handling panics
    panic::set_hook(Box::new(|panic_info| {
        if let Some(payload) = panic_info.payload().downcast_ref::<&str>() {
            error!("{}", payload);
        } else {
            error!("{}", panic_info);
        }
    }));

    let patterns = Patterns::new(&config.scanner);
    let providers = Providers::new();
    let scanner = Scanner::new(&config.scanner, &providers, &patterns);

    for request in Listener::new() {
        let result = scanner.scan(&request);
        println!("{}", serde_json::to_string(&result)?);
    }

    Ok(())
}
