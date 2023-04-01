use std::io::Lines;
use std::io::{self, StdinLock};
use std::iter::Iterator;

use crate::scanner::proto::Request;

struct Requests {
    lines: Lines<StdinLock<'static>>,
}

impl Requests {
    fn next(&mut self) -> Option<Request> {
        self.lines
            .next()
            .and_then(Result::ok)
            .and_then(|s| serde_json::from_str(&s).ok())
    }
}

pub struct Listener {
    requests: Requests,
}

impl Default for Listener {
    fn default() -> Self {
        Self::new()
    }
}

impl Listener {
    pub fn new() -> Listener {
        Listener {
            requests: Requests {
                lines: io::stdin().lines(),
            },
        }
    }
}

impl Iterator for Listener {
    type Item = Request;

    fn next(&mut self) -> Option<Self::Item> {
        self.requests.next()
    }
}
