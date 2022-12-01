use serde::{self, Deserialize, Serialize};

// Options for a git scan
#[derive(Debug, Deserialize)]
pub struct GitOptions {
    // Set --depth for the git clone
    pub depth: Option<u32>,
}

// The incoming request
#[derive(Debug, Deserialize)]
#[serde(tag = "type")]
pub enum Request {
    #[serde(rename = "git")]
    Git {
        // A way to tie the response to the request
        id: String,
        // The URL of the repo to scan
        url: String,
        // Optional options for how the git scan
        options: Option<GitOptions>,
    },
}

// The fields from the Request that should be included in response.request
#[derive(Debug, Serialize)]
pub struct ResponseRequest {
    // A way to tie the response to the request
    pub id: String,
}

// GitleaksRule details
#[derive(Debug, Serialize)]
pub struct Rule {
    // The unique rule id
    pub id: String,
    // A description of what the rule finds
    pub description: String,
    // A set of tags that tells the scanner what to do with the result
    pub tags: Vec<String>,
}

// The section of a file where a scan result was found
#[derive(Debug, Serialize)]
pub struct Lines {
    // The start line
    pub start: u32,
    // The end line
    pub end: u32,
}

// The comitter's details
#[derive(Debug, Serialize)]
pub struct GitCommitAuthor {
    // The author's name
    pub name: String,
    // The author's email
    pub email: String,
}

// The commit's details
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

// Details about the source where the result was found
#[derive(Debug, Serialize)]
#[serde(tag = "type")]
pub enum Source {
    // When the scan source is a git repo
    #[serde(rename = "git")]
    Git {
        // The URL to the remote or local git repo
        url: String,
        // The path to the result relative to the repo
        path: String,
        // The line range for the result
        lines: Lines,
        // Info about the commit containing the result
        commit: GitCommit,
    },
}

#[derive(Debug, Serialize)]
pub struct Result {
    // The match that triggered the rule
    pub context: String,
    // The specific thing the rule was meant to find
    pub target: String,
    // The entropy associated result
    pub entropy: f32,
    // Which rule triggered the result
    pub rule: Rule,
    // Where the result was found
    pub source: Source,
}

// Something meant to hold data from a gitleaks scan to map it to a Result
// see the gitleaks docs for the meaning of each field
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

// The full response that includes all of the meta data and results
#[derive(Debug, Serialize)]
pub struct Response {
    // A unique id generated for each scan
    pub id: String,
    // Details from the request so the response can be tied back to the
    // original request
    pub request: ResponseRequest,
    // The individual results of the scan
    pub results: Vec<Result>,
}
