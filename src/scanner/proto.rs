use serde::{self, Deserialize, Serialize};

#[derive(Debug, Serialize, Deserialize)]
pub enum Kind {
    #[serde(rename = "git")]
    Git,
    // Some future ideas:
    // * EmailAddress - for pattern matches + HIBP breach checks
    // * JiraTicketURL
    // * String or some kind of raw content to be scanned directly
}

#[derive(Debug, Deserialize)]
pub struct GitOptions {
    pub clone_depth: Option<u32>,
}

// The incoming request
#[derive(Debug, Deserialize)]
#[serde(tag = "type")]
pub enum Request {
    #[serde(rename = "git")]
    Git {
        // A way to identify the response to the request
        id: String,
        // What kind of thing is being scanned
        // The thing that should be scanned
        url: String,
        // Scanner options for the specific kind of url being scanned
        options: Option<GitOptions>,
    },
}

// The fields from a request that should be returned in the response
// See Request for the meaning of the fields
#[derive(Debug, Serialize)]
pub struct ResponseRequest {
    pub id: String,
}

#[derive(Debug, Serialize)]
pub struct Rule {
    // The unique rule id
    pub id: String,
    // A description of the rule and what it finds
    pub description: String,
    // A set of standardized tags
    pub tags: Vec<String>,
}

#[derive(Debug, Serialize)]
pub struct Lines {
    // The start a section of a source
    pub start: u32,
    // The end a section of a source
    pub end: u32,
}

#[derive(Debug, Serialize)]
pub struct GitCommitAuthor {
    // The author's name from a git commit
    pub name: String,
    // The author's email from a git commit
    pub email: String,
}

#[derive(Debug, Serialize)]
pub struct GitCommit {
    // The sha1 to identify a commit
    pub id: String,
    // The author of the commit
    pub author: GitCommitAuthor,
    // When the commit was created
    pub date: String,
    // The commit message
    pub message: String,
}

#[derive(Debug, Serialize)]
#[serde(tag = "type")]
pub enum Source {
    // When the scan source is a git repo
    #[serde(rename = "git")]
    Git {
        // The URL to the remote or local git repo
        url: String,
        // The path to the leak relative to the repo
        path: String,
        // The line range for the leak
        lines: Lines,
        // Info about the commit containing the leak
        commit: GitCommit,
    },
}

#[derive(Debug, Serialize)]
pub struct Result {
    // The context around the target that the rule triggered on
    pub context: String,
    // The specific thing the rule was meant to find
    pub target: String,
    // The entropy associated with the match
    pub entropy: f32,
    // Which rule triggered the match
    pub rule: Rule,
    // Where the result was found
    pub source: Source,
}

#[derive(Debug, Deserialize)]
pub struct GitLeaksResult {
    #[serde(rename = "Match")]
    pub context: String,
    #[serde(rename = "Secret")]
    pub target: String,
    #[serde(rename = "Entropy")]
    pub entropy: f32,
    #[serde(rename = "RuleID")]
    pub rule_id: String,
    #[serde(rename = "Description")]
    pub rule_description: String,
    #[serde(rename = "Tags")]
    pub rule_tags: Vec<String>,
    #[serde(rename = "File")]
    pub source_path: String,
    #[serde(rename = "StartLine")]
    pub source_lines_start: u32,
    #[serde(rename = "EndLine")]
    pub source_lines_end: u32,
    #[serde(rename = "Commit")]
    pub source_commit_id: String,
    #[serde(rename = "Date")]
    pub source_commit_date: String,
    #[serde(rename = "Message")]
    pub source_commit_message: String,
    #[serde(rename = "Author")]
    pub source_commit_author_name: String,
    #[serde(rename = "Email")]
    pub source_commit_author_email: String,
}

#[derive(Debug, Serialize)]
pub struct Response {
    pub id: String,
    pub request: ResponseRequest,
    pub results: Vec<Result>,
}
