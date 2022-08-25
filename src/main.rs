pub mod config;
pub mod scanner;

use crate::config::Config;
use crate::scanner::patterns;

fn main() {
    // TODO: move this code out of here
    // this is just stubbed out as things are getting set up
    let config = Config::load(r#"
        [scanner]
        workdir = "/tmp/leaktk"

        [scanner.patterns]
        server_url = "https://raw.githubusercontent.com/leaktk/patterns/main/target/patterns"
        refresh_interval = 43200
    "#);

    patterns::refresh(&config.scanner);
    println!("{:#?}", config);
}
