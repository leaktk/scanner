use crate::scanner::proto::Request;
use std::io::Lines;
use std::io::{self, StdinLock};
use std::iter::Iterator;

struct Requests {
    lines: Lines<StdinLock<'static>>,
}

impl Requests {
    fn next(&mut self) -> Option<Request> {
        self.lines
            .next()
            .map(Result::ok)
            .flatten()
            .map(|s| serde_json::from_str(&s).ok())
            .flatten()
    }
}

pub struct Listner {
    requests: Requests,
}

impl Listner {
    pub fn new() -> Listner {
        Listner {
            requests: Requests {
                lines: io::stdin().lines(),
            },
        }
    }
}

impl Iterator for Listner {
    type Item = Request;

    fn next(&mut self) -> Option<Self::Item> {
        self.requests.next()
    }
}
