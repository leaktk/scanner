# Scanner

Provides a consistent API around some existing scanning tools to integrate them
with the rest of the tool kit.

The scanner can run in listening mode to handle a stream of requests or it can
run single scans.

The scanner leverages
[Gitleaks](https://github.com/gitleaks/gitleaks)
internally because Gitleaks is an awesome tool, and we already have quite a few
[patterns built up](https://github.com/leaktk/patterns)
for it.

## Status

Updating patterns to support this version.

## Usage

```sh
# Listen for requests
leaktk-scanner listen < ./examples/requests.jsonl

# Run a single scan
leaktk-scanner scan --resource 'https://github.com/leaktk/fake-leaks.git'
leaktk-scanner scan --kind JSONData --resource '{"key": "-----BEGIN PRIVATE KEY-----c5602d28d0f21422dfc7b572b17e6b138c1b49fd7f477d4c5c961e0756f1ff70-----END PRIVATE KEY-----"}'

# See more options
leaktk-scanner help
```

When in listening mode the scanner listens on stdin, responds on stdout, and
logs to stderr. It reads one request per line and sends a response per line in
jsonl.

The "Scan Request Format" and "Scan Results Format" sections describe the
format for requests and responses.

## Config File Format

The [config.go](./pkg/config/config.go) file sets the default config values and
[examples/config.toml](./examples/config.toml) contains a commented version of
the config file explaining what the settings mean.

The order of precedence for config paths:

1. The `LEAKTK_CONFIG_PATH` env var
1. `--config <some path>`
1. `${XDG_CONFIG_HOME}/leatktk/config.toml` if it exists
1. `/etc/leaktk/config.toml` if it exists
1. default config

There are also several env vars that take precedence over the other config
settings if they're set:

- `LEAKTK_LOGGER_LEVEL` - set the level of the logger
- `LEAKTK_PATTERN_SERVER_AUTH_TOKEN` - the pattern server token
- `LEAKTK_PATTERN_SERVER_URL` - the pattern server URL
- `LEAKTK_SCANNER_AUTOFETCH` - whether the scanner can auto fetch patterns or
  any other items it may need to do the scan (except for the resource being
  scanned)

## Request/Response formats

Notes about the formats below:

* Scan requests should be sent as [JSON lines](https://jsonlines.org/).
* The requests below are pretty printed to make them easier to read.
* Only the values in the `"options"` sections are optional.

### GitRepo

#### Request

```json
{
  "id": "db0c21127a6a849fdf8eeae65d753275f3a26a33649171fa34af458030744999",
  "kind": "GitRepo",
  "resource": "https://github.com/leaktk/fake-leaks.git",
  "options": {
    "branch": "main",
    "depth": 1,
    "proxy": "http://squid.example.com:3128",
    "since": "2020-01-01",
  }
}
```

The `options` above are not required and some combined (e.g. `depth` and `since`)
may cause issues. Refer to the details below to better understand the options
and [git's docs](https://git-scm.com/) for knowing what can be combined.

#### Request Options

**branch**

Sets `--branch` and `--single-branch` during git clone.

* Type: `string`
* Default: excluded

**depth**

Sets `--depth` during a git clone and can limit the commits during a local
scan if `single_branch` is set to true.

* Type: `uint16`
* Default: excluded

**since**

Is a date formatted `yyyy-mm-dd` used for filtering commits. Sets
`--shallow-since` during a clone.

* Type: `string`
* Default: excluded

**proxy**

A URL for a http proxy. Sets `--config http.proxy` during the
clone.

* Type: `string`
* Default: excluded

#### Response

```json
{
  "id": "8343516f29a9c80cc7862e01799f446d5fb93088d1681f8c5181b211488a94db",
  "request": {
    "id": "6d56db314f87371d0c2da1b1bcf90c9594e0bf793280e6ddd896d69881e6099b",
    "kind": "GitRepo",
    "resource": "http://github.com/leaktk/fake-leaks.git"
  },
  "results": [
    {
      "id": "c80efb11ee9013a42d6037566c7159c9aeb610696be851dc9209c85e75e5a3e7",
      "kind": "GitCommit",
      "secret": "-----BEGIN EC PRIVATE KEY-----\n...snip...\n-----END EC PRIVATE KEY-----",
      "match": "-----BEGIN EC PRIVATE KEY-----\n...snip...\n-----END EC PRIVATE KEY-----",
      "entropy": 5.763853,
      "date": "2023-08-04T12:21:12Z",
      "rule": {
        "id": "private-key",
        "description": "Private Key",
        "tags": [
          "group:leaktk-testing",
          "alert:repo-owner",
          "type:secret"
        ]
      },
      "contact": {
        "name": "The Committer",
        "email": "user@example.com"
      },
      "location": {
        "version": "d5bcb89de5311aaa688cb23d8d2d78cf7cd74f1f",
        "path": "keys/tls/another-key.key",
        "start": {
          "line": 1,
          "column": 1
        },
        "end": {
          "line": 6,
          "column": 29
        }
      },
      "notes": {
        "message": "Add another key"
      }
    }
  ]
}
```


### JSONData

#### Request

```json
{
  "id": "580efb11ee9013a42d6037566c7159c9aeb610696be851dc9209c85e75e5a3e7",
  "kind": "JSONData",
  "resource": "{\"some\": {\"key\": \"-----BEGIN PRIVATE KEY-----c5602d28d0f21422dfc7b572b17e6b138c1b49fd7f477d4c5c961e0756f1ff70-----END PRIVATE KEY-----\"}}"
}
```

Note: the `resource` here is a string containing escaped JSON data

#### Request Options

JSONData currently doesn't have any options but it does support including
Gitleaks config files (e.g. `.gitleaks.toml`, `.gitleaksignore`,
`.gitleaksbaseline`) as top level keys that will be considered during the
scan.

#### Response

```json
{
  "id": "c88057777737115851ad94b91461d09b2ce704484813557895ba0d6d827d4ed8",
  "request": {
    "id": "580efb11ee9013a42d6037566c7159c9aeb610696be851dc9209c85e75e5a3e7",
    "kind": "JSONData",
    "resource": "{\"some\":{\"key\": \"-----BEGIN PRIVATE KEY-----c5602d28d0f21422dfc7b572b17e6b138c1b49fd7f477d4c5c961e0756f1ff70-----END PRIVATE KEY-----\"}}"
  },
  "results": [
    {
      "id": "66456b43e1efac03f9448ece59a65b0e2bf304b55506507a8aa07727e3900522",
      "kind": "JSONData",
      "secret": "-----BEGIN PRIVATE KEY-----c5602d28d0f21422dfc7b572b17e6b138c1b49fd7f477d4c5c961e0756f1ff70-----END PRIVATE KEY-----",
      "match": "-----BEGIN PRIVATE KEY-----c5602d28d0f21422dfc7b572b17e6b138c1b49fd7f477d4c5c961e0756f1ff70-----END PRIVATE KEY-----",
      "entropy": 4.490795,
      "date": "",
      "rule": {
        "id": "private-key",
        "description": "Private Key",
        "tags": [
          "group:leaktk-testing",
          "alert:repo-owner",
          "type:secret"
        ]
      },
      "contact": {
        "name": "",
        "email": ""
      },
      "location": {
        "version": "",
        "path": "some/key",
        "start": {
          "line": 0,
          "column": 1
        },
        "end": {
          "line": 0,
          "column": 116
        }
      },
      "notes": {
        "message": ""
      }
    }
  ]
}
```

Note: the `path` here is formatted like a file path. For arrays, the element's
index is used.

## TODO

1. Finish and test steps to make it usable inside `leaktk-monitor` (aka PwnedAlert)
1. Make sure features in the internal scanner are upstreamed here
1. Support local scans
1. Delete existing rust projects from crates.io
1. Make sure it fully supports Linux and Mac
1. Start integrating this into a pre-commit hooks project
1. Figure out a way to apply different rules in different contexts (internal/external repos, etc)
1. Support `.github/secret_scanning.yml` files
1. Explore base64 support (i.e. decoding it when spotted and scanning the contents)
