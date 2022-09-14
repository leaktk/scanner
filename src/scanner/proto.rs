use serde::{self, Deserialize, Serialize};

#[derive(Debug, Serialize, Deserialize)]
pub enum Kind {
    GitRepoURL,
    // Some future ideas:
    // * EmailAddress
    // * JiraTicketURL
}

#[derive(Debug, Serialize, Deserialize)]
pub struct Request<'r> {
    pub kind: Kind,
    pub artifact: &'r str,
}

#[derive(Debug)]
pub struct Response<'r> {
    pub request: &'r Request<'r>,
}

impl<'r> Response<'r> {
    pub fn new(req: &'r Request) -> Response<'r> {
        Response { request: req }
    }
}
