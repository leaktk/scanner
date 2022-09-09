pub mod config;
pub mod scanner;

use crate::config::Config;
use crate::scanner::{proto::Kind, proto::Request, Scanner};

fn main() {
    // TODO: move this code out of here
    // this is just stubbed out as things are getting set up
    let config = Config::load(
        r#"
        [scanner]
        workdir = "/tmp/leaktk"

        [scanner.patterns]
        server_url = "https://raw.githubusercontent.com/leaktk/patterns/main/target/patterns"
        refresh_interval = 43200
    "#,
    );

    let mut scanner = Scanner::new(&config.scanner);

    // TODO: wrap this in an io handler for different methods
    let reqs = vec![
        Request {
            kind: Kind::Git,
            url: "https://github.com/leaktk/fake-leaks.git",
        },
        Request {
            kind: Kind::Git,
            url: "https://github.com/leaktk/fake-leaks.git",
        },
    ];

    for req in &reqs {
        for resp in &scanner.scan(&req) {
            // TODO: wrap this in an io handler for different methods
            println!("{:#?}", resp);
        }
    }
}
