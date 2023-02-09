# Scanner

Provides a consistent API around some existing scanning tools to integrate them
with the rest of the tool kit.

This scanner is meant to be ran as a single instance listening on stdin
for easy scripting. There will be a server to wrap the scanner to be able to
turn it into a service

The scanner leverages
[gitleaks](https://github.com/zricethezav/gitleaks)
internally because gitleaks is an awesome tool, and we already have quite a few
[patterns built up](https://github.com/leaktk/patterns)
for it.

## Status

Just getting started.

## Usage

```sh
leaktk-scanner --config ./examples/config.toml < ./examples/requests.jsonl
```

The scanner listens on stdin, responds on stdout, and logs to stderr.
It reads one request per line and sends a response per line in jsonl.

The "Scan Request Format" and "Scan Results Format" sections describe the
format for requests and responses.

## Config File Format

The [example config](./examples/config.toml) has the default values set and
comments explaining what each option does.

## Scan Request Format

Notes about the formats below:

* Scan requests should be sent as JSON Lines.
* The requests below are pretty printed to make them easier to read.
* Only the values in the `"options"` sections are optional.

### Git

Scan git repos

```json
{
  "id": "1bc1dc91-3699-41cf-9486-b74f0897ae4c",
  "kind": "git",
  "target": "https://github.com/leaktk/fake-leaks.git",
  "options": {
      "local": false,
      "branch": "main",
      "depth": 5,
      "since": "2020-01-01",
      "single_branch": true,
      "config": [
        "http.sslVerify=true"
      ]
  }
}
```

Supported options:

* Git options
    * `config:Vec<String>` -> `[--config String ...]`
    * `since:String` -> `--shallow-since String`
    * `single_branch:bool` -> `--[no-]single-branch`
    * `depth:u32` -> `--depth u32`
    * `branch:String` -> `--branch String`
* Scanner options
    * `local:bool` - Skip clone and target like a path

The git options will be passed to the git command, even if the
combination of options doesn't make sense.

When `local` is set to true (it defaults to false),

* `target` will be interpreted as a path.
* The following options will be ignored:
    * config

**WARNING**: The `local` option should be striped from the requests if passing
untrusted input to the scanner.

## Scan Results Format

The scan result format is in jsonl here are formatted examples of a single
line by kind.

### Git

Success

```json
{
  "id": "dd4f7ac3-134f-489a-b0a9-0830ab98e271",
  "request": {
    "id": "1bc1dc91-3699-41cf-9486-b74f0897ae4c"
  },
  "results": [
    {
      "context": "<the match>",
      "target": "<the part of the match the rule was trying to find>",
      "entropy": 3.2002888,
      "rule": {
        "id": "some-rule-id",
        "description": "A human readable description of the rule",
        "tags": [
          "alert:repo-owner",
          "type:secret",
        ]
      },
      "source": {
        "kind": "git",
        "target": "https://github.com/leaktk/fake-leaks.git",
        "path": "relative/path/to/the/file",
        "lines": {
          "start": 1,
          "end": 1
        },
        "commit": {
          "id": "<commit-sha>",
          "author": {
            "name": "John Smith",
            "email": "jsmith@example.com"
          },
          "date": "2022-08-29T15:32:48Z",
          "message": "<commit message>"
        }
      }
    }
  ]
}
```

Error (if "error" is present, the scan failed)

```json
{
  "id": "57920e78-b89d-4dbd-ac59-8d9750eb0515",
  "request": {
    "id": "1bc1dc91-3699-41cf-9486-b74f0897ae4c"
  },
  "error": "<some error message>",
  "results": []
}
```

## TODO

1. Local git scans without a clone
1. Support unstaged commits (probably via gitleaks protect)
1. Better error handling in the code to avoid panics
1. Sanitize any repo specific .gitleaks.tomls and load them as a part of the scans
1. Change workdir default to `${XDG_CACHE_HOME}/leaktk`
1. if --config isn't set:
    * then try `${XDG_CONFIG_HOME}/leatktk/config.toml`
    * else try /etc/leaktk/config.toml
    * else fall back on default config
1. Allow optional config headers passed to the pattern server requests
1. Finalize the request/response format
1. Make sure it fully supports Linux and Mac
1. Unittest and refactor what's currently here
1. Proper error handling in the code to keep things clean, consistent and scalable
1. Group gitleaks code into a single object as the source of truth
1. Create a Workspace object to manage the workspace folders (creating, clearing, etc)
1. Figure out what to do with shallow commits on shallow-since scans
1. Look into creating rust bindings to call gitleaks directly from rust instead of spinning up a process
