pub mod config;
pub mod scanner;

use std::fs;
use crate::config::Config;
use crate::scanner::Scanner;

fn main() {
    // TODO: move this code out of here
    // this is just stubbed out as things are getting set up
    let config = Config::load(&fs::read_to_string("./config.toml").unwrap());
    let raw_requests = fs::read_to_string("./reqs.jsonl").unwrap();

    let mut scanner = Scanner::new(&config.scanner);
    for line in raw_requests.lines() {
        for resp in &scanner.scan(&serde_json::from_str(line).unwrap()) {
            // TODO: wrap this in an io handler for different methods
            println!("{:#?}", resp);
        }
    }
}
