use std::collections::HashMap;
use std::env;
use std::process;

pub struct Args {
    pub config: Option<String>,
}

fn parse<I: Iterator<Item = String>>(
    mut args: I,
    mut map: HashMap<String, String>,
) -> HashMap<String, String> {
    loop {
        if let Some(arg) = args.next() {
            if arg.starts_with("--") {
                if let Some(value) = args.next() {
                    map.insert(arg[2..].to_string(), value);
                } else {
                    eprintln!("The option \"{}\" was missing a value.\n", arg);
                    eprintln!("USAGE\n\n    leaktk-scanner --config CONFIG_PATH");
                    process::exit(1);
                }
            }
        } else {
            break map;
        }
    }
}

pub fn args() -> Args {
    let raw_args = parse(env::args(), HashMap::new());

    Args {
        config: raw_args.get("config").map(|s| s.to_string()),
    }
}
