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
# TODO: make --config optional and if config isn't set:
# 1) then try ${XDG_CONFIG_HOME}/leatktk/config.toml
# 2) else try /etc/leaktk/config.toml
# 3) else fall back on default config
leaktk-scanner --config ./examples/config.toml < ./examples/requests.jsonl
```

The scanner listens on stdin, responds on stdout, and logs to stderr.
It reads one request per line and sends a response per line in jsonl.

The "Scan Request Format" and "Scan Results Format" sections describe the
format for requests and responses.

## Config File Format

The [example config](./examples/config.toml) provides a commented version with
all of the default values set.

## Scan Request Format

The scan request format is in jsonl here are formatted examples of a single
line by type. Only the values in the `"options"` sections are optional.

**WARNING**: Certain request types (e.g. `"type": "git", "url": "file://..."`)
can access files outside of the scanner's workdir. Make sure you trust or
sanitize the input to the scanner.

### Git (Remote)

Clone a remote repo and scan it.

```json
{
  "id": "<uuid>",
  "type": "git",
  "url": "https://github.com/leaktk/fake-leaks.git",
  "options": {
    "depth": 5,
  }
}
```

### TODO: Git (Local)

Scan a local repo. Instead of cloning the repo, the scanner will simply
scan the contents of the existing repo. This can be useful for implementing
pre-commit hooks or tool-chains that already take care of cloning the repo.

```json
{
  "id": "<uuid>",
  "type": "git",
  "url": "file:///home/user/workspace/leaktk/fake-leaks",
  "options": {
    "depth": 5,
  }
}
```

## Scan Results Format

The scan result format is in jsonl here are formatted examples of a single
line by type

### Git

```json
{
  "id": "<uuid>",
  "request": {
    "id": "<uuid>"
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
        "type": "git",
        "url": "https://github.com/leaktk/fake-leaks.git",
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

## TODO

* Fix/cleanup the error handling
* Proper logging
* All the TODOs called out in the README
* Unittest and refactor what's currently there
* Group gitleaks code into a single object as the source of truth
* Create a Workspace object to manage the workspace folders (creating, clearing, etc)
* Encapsulate some of the Linux specific bits
