use crate::config::{ListnerConfig, ListnerMethod};
use std::io::Lines;
use std::io::{self, StdinLock};
use std::iter::Iterator;

enum Requests {
    Stdin(Lines<StdinLock<'static>>),
}

impl Requests {
    fn next(&mut self) -> Option<String> {
        match self {
            Self::Stdin(lines) => lines.next().map(Result::ok).flatten(),
        }
    }
}

pub struct Listner {
    requests: Requests,
}

impl Listner {
    pub fn new(config: &ListnerConfig) -> Listner {
        Listner {
            requests: match config.method {
                ListnerMethod::Stdin => Requests::Stdin(io::stdin().lines()),
            },
        }
    }
}

impl Iterator for Listner {
    type Item = String;

    fn next(&mut self) -> Option<Self::Item> {
        self.requests.next()
    }
}
