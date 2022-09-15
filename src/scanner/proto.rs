use std::collections::HashMap;
use serde::{self, Deserialize, Serialize};

#[derive(Debug, Serialize, Deserialize)]
pub enum Kind {
    GitRepoURL,
    // Some future ideas:
    // * EmailAddress - for pattern matches + HIBP breach checks
    // * JiraTicketURL
    // * String or some kind of raw content to be scanned directly
}

#[derive(Debug, Serialize, Deserialize)]
pub struct Request {
    pub kind: Kind,
    pub artifact: String,
    pub options: Option<HashMap<String,String>>,
}

#[derive(Debug)]
pub struct Response<'r> {
    pub request: &'r Request,
}

impl<'r> Response<'r> {
    pub fn new(req: &'r Request) -> Response<'r> {
        Response { request: req }
    }
}
