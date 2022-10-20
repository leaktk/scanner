pub mod config;
pub mod listner;
pub mod scanner;

use crate::config::Config;
use crate::listner::Listner;
use crate::scanner::Scanner;
use std::env;
use std::fs;

fn main() {
    let config_path = env::args()
        .nth(1)
        .expect("First argument must be the config path");

    let config = Config::load(&fs::read_to_string(config_path).expect("Unable to load config"));
    let mut scanner = Scanner::new(&config.scanner);

    for request in Listner::new(&config.listner) {
        for resp in &scanner.scan(&request) {
            // TODO: wrap this in an io handler for different methods
            println!("{}", serde_json::to_string(&resp).unwrap());
        }
    }
}
