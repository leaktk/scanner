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

The order of precedence for config paths:

1. `--config <some path>`
1. `${XDG_CONFIG_HOME}/leatktk/config.toml` if it exists
1. `/etc/leaktk/config.toml` if it exists
1. default config

## Scan Request Format

Notes about the formats below:

* Scan requests should be sent as JSON Lines.
* The requests below are pretty printed to make them easier to read.
* Only the values in the `"options"` sections are optional.

**WARNING**: Sanitize the input before passing it to the scanner. Work will
be done to harden it a bit more, but don't run this as something taking user
input on a host running this as a service.

### Git

Scan git repos

```json
{
  "id": "1bc1dc91-3699-41cf-9486-b74f0897ae4c",
  "kind": "git",
  "target": "https://github.com/leaktk/fake-leaks.git",
  "options": {
      "branch": "main",
      "config": ["http.sslVerify=true"],
      "depth": 5,
      "since": "2020-01-01",
      "single_branch": true
  }
}
```

#### Options

**branch**

Sets `--branch` during git clone. If you wish to only scan this branch in a
local scan, set `single_branch` to true as well.

* Type: `String`
* Default: Excluded
* Supported by local scan: yes
* Supported by remote scan: yes

**config**

Is a list of key=value strings that get passed to git using the `--config`
flag.

* Type: `Vec<String>`
* Default: Excluded
* Supported by local scan: no
* Supported by remote scan: yes

**depth**

Sets `--depth` during a git clone and can limit the commits during a local
scan if `single_branch` is set to true.

* Type: `u32`
* Default: Excluded
* Supported by local scan: partial
* Supported by remote scan: yes

**local**

Skips the clone and `target` is treated as a path.

* Type: `bool`
* Default: `false`
* Defines if it is a local or remote scan

**since**

Is a date formatted `yyyy-mm-dd` used for filtering commits.

* Type: `String`
* Default: Excluded
* Supported by local scan: yes
* Supported by remote scan: yes

**single_branch**

Sets the branch to clone and the scope of the gitleaks scan.

* Type: `bool`
* Default: `false`
* Supported by local scan: yes
* Supported by remote scan: yes

**staged**

Scan staged changes (implies `uncommitted` and is useful for pre-commit hooks).

* Type: `bool`
* Default: `false`
* Supported by local scan: yes
* Supported by remote scan: no

**uncommitted**

Scan uncommitted changes (implied by `staged`)

* Type: `bool`
* Default: `false`
* Supported by local scan: yes
* Supported by remote scan: no

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

1. Fix clippy warnings and add it to the contributing guidelines
1. Allow optional config headers passed to the pattern server requests
1. Finalize the request/response format
1. Make sure it fully supports Linux and Mac
1. Unittest and refactor what's currently here
1. Proper error handling in the code to keep things clean, consistent and scalable
1. Group gitleaks code into a single object as the source of truth
1. Figure out what to do with shallow commits on shallow-since scans
1. Look into creating rust bindings to call gitleaks directly from rust instead of spinning up a process
1. Figure out a fast way for depth limiting when single\_branch is set to false where it gives n commits from each branch
1. Figure out a way to apply different rules in different contexts (internal/external repos, etc)
1. Support `.github/secret_scanning.yml` files
1. Explore base64 support (i.e. decoding it when spotted and scanning the contents)
