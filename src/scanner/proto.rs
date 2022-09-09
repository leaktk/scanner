use serde::{self, Deserialize, Serialize};

#[derive(Debug, Serialize, Deserialize)]
pub enum Kind {
    Git,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct Request<'r> {
    pub kind: Kind,
    pub url: &'r str,
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
