# Scanner

Provides a consistent API around some existing scanning tools to integrate them
with the rest of the tool kit.

The scanner can run in listening mode to handle a stream of requests or it can
run single scans.

The scanner leverages
[Gitleaks](https://github.com/gitleaks/gitleaks)
internally because Gitleaks is an awesome tool, and we already have quite a few
[patterns](https://github.com/leaktk/patterns)
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

When in `listen` mode, the scanner listens on stdin, responds on stdout, and
logs to stderr. It reads one request per line and sends one response per line
in jsonl. The scanner should always generate a response to each request even if
there were errors during the scan.

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
* The examples below are pretty printed to make them easier to read.
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
    "id": "db0c21127a6a849fdf8eeae65d753275f3a26a33649171fa34af458030744999",
    "kind": "GitRepo",
    "resource": "http://github.com/leaktk/fake-leaks.git"
  },
  "errors": [],
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

This allows you to scan various JSON structures for secrets.

#### Request

```json
{
  "id": "580efb11ee9013a42d6037566c7159c9aeb610696be851dc9209c85e75e5a3e7",
  "kind": "JSONData",
  "resource": "{\"some\": {\"key\": \"-----BEGIN PRIVATE KEY-----c5602d28d0f21422dfc7b572b17e6b138c1b49fd7f477d4c5c961e0756f1ff70-----END PRIVATE KEY-----\"}}"
}
```

Note: the `resource` here is a string containing escaped JSON data.

#### Request Options

**fetch_urls**

If true and one of the values in the JSON data is a URL, it will be fetched
and treated as the content at that path in the data structure. This is
like running a URL scan. If the response is JSON it can be traversed too,
but it won't fetch futher URLs.

* Type: `bool`
* Default: `false`

**NOTE:** It is **NOT** recommended to enable this option unless you trust
the JSON and the URLs you are providing to the scanner.

Gitleaks config files (e.g. `.gitleaks.toml`, `.gitleaksignore`,
`.gitleaksbaseline`) as top level keys that will be considered during the
scan. The [examples/requests.jsonl](./examples/requests.jsonl) has an
example of including a `.gitleaks.toml` with a JSONData scan.

#### Response

```json
{
  "id": "c88057777737115851ad94b91461d09b2ce704484813557895ba0d6d827d4ed8",
  "request": {
    "id": "580efb11ee9013a42d6037566c7159c9aeb610696be851dc9209c85e75e5a3e7",
    "kind": "JSONData",
    "resource": "{\"some\":{\"key\": \"-----BEGIN PRIVATE KEY-----c5602d28d0f21422dfc7b572b17e6b138c1b49fd7f477d4c5c961e0756f1ff70-----END PRIVATE KEY-----\"}}"
  },
  "errors": [],
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
      "notes": {}
    }
  ]
}
```

Note: the `path` here is formatted like a file path. For arrays, the element's
index is used.

### Files

This allows you to scan files and directories.

#### Request

```json
{
  "id": "1c7387179582ae1e9bc114bfa10bddc6317fe6a5362efd2ae4019e34cccd8420",
  "kind": "Files",
  "resource": "/path/to/fake-leaks/keys/tls"
}
```

Note: the `resource` can be either a single file or a directory.

#### Request Options

Files currently doesn't have any options but all the Gitleaks config files
(e.g. `.gitleaks.toml`, `.gitleaksignore`, `.gitleaksbaseline`) are supported.

#### Response

```json
{
  "id": "bde3a15ca81cf6503b9c9e1a450a3bbdfa09567e77f787a6cf4a56ed3b115f87",
  "request": {
    "id": "1c7387179582ae1e9bc114bfa10bddc6317fe6a5362efd2ae4019e34cccd8420",
    "kind": "Files",
    "resource": "/path/to/fake-leaks/keys/tls"
  },
  "errors": [],
  "results": [
    {
      "id": "9376e604a7e8c4f5259ceb47f8a29c57e77c07668e317fa8c177de5e56fbe029",
      "kind": "General",
      "secret": "-----BEGIN EC PRIVATE KEY-----\n...snip...\n-----END EC PRIVATE KEY-----",
      "match": "-----BEGIN EC PRIVATE KEY-----\n...snip...\n-----END EC PRIVATE KEY-----",
      "entropy": 5.763853,
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
        "path": "another-key.key",
        "start": {
          "line": 0,
          "column": 1
        },
        "end": {
          "line": 5,
          "column": 29
        }
      },
      "notes": {}
    }
  ]
}
```

Note: the `path` is relative to the resource provided. If the resource is the
path to the file itself, then path will be empty.

### URL

This allows you to pull remote content to scan. It also has basic awareness
of the response `Content-Type`. If the response has a content type of
`application/json` it will be parsed as a `JSONData` request. Else it will
be parsed as a `Files` request.

#### Request

```json
{
  "id": "1c7387179582ae1e9bc114bfa10bddc6317fe6a5362efd2ae4019e34cccd8420",
  "kind": "URL",
  "resource": "https://raw.githubusercontent.com/leaktk/fake-leaks/main/keys/tls/server.key"
}
```

#### Request Options

URL currently doesn't have any options.

#### Response

```json
{
  "id": "a1ef32d00c609b370d2181ea46babbbdd19deeeea68918cc676a8f12d1fc7e3b",
  "request": {
    "id": "1c7387179582ae1e9bc114bfa10bddc6317fe6a5362efd2ae4019e34cccd8420",
    "kind": "URL",
    "resource": "https://raw.githubusercontent.com/leaktk/fake-leaks/main/keys/tls/server.key"
  },
  "errors": [],
  "results": [
    {
      "id": "6c14f496a2111dfeecbfff4a61587b0b1866788a6112b420f80071f8cded0153",
      "kind": "General",
      "secret": "-----BEGIN PRIVATE KEY-----\n...snip...\n-----END PRIVATE KEY-----",
      "match": "-----BEGIN PRIVATE KEY-----\n...snip...\n-----END PRIVATE KEY-----",
      "entropy": 6.0285063,
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
        "path": "",
        "start": {
          "line": 2,
          "column": 2
        },
        "end": {
          "line": 29,
          "column": 26
        }
      },
      "notes": {}
    }
  ]
}
```

Note: the `path` will be blank when using a `Files` type scan and it will
use the same logic as `JSONData` when the response's content type is
`application/json` (i.e. it will be the path down the traversed keys).

### Container Image

This allows you to pull a remote container image to scan. It unpacks and scans
the Image, Config and Manifest.

#### Request

```json
{
  "id": "1c7387179582ae1e9bc23123a10bddc6317fe6a5362efd2ae4019e34cccd8420",
  "kind": "ContainerImage",
  "resource": "quay.io/wizzy/fake-leaks:v1.0.2"
}
```

#### Request Options

**exclusions**

Sets a list of RootFS Layer hashes to exclude from scanning

* Type: `[]string`
* Default: excluded

**arch**

Provide a preferred architecture

* Type: `string`
* Default: excluded

#### Response  
TODO: Refine this.
```json
{
  "id": "a1ef32d00c609b370d2181ea46b11111119deeeea68918cc676a8f12d1fc7e3b",
  "request": {
    "id": "1c7387179582ae1e9bc23123a10bddc6317fe6a5362efd2ae4019e34cccd8420",
    "kind": "ContainerImage",
    "resource": "quay.io/wizzy/fake-leaks:v1.0.2"
  },
  "results": [
    {
      "id": "6c14f496a2111dfeecbfff4a61587b0b1866788a6112b420f80071f8cded0153",
      "kind": "ContainerLayer",
      "secret": "-----BEGIN PRIVATE KEY-----\n...snip...\n-----END PRIVATE KEY-----",
      "match": "-----BEGIN PRIVATE KEY-----\n...snip...\n-----END PRIVATE KEY-----",
      "entropy": 6.0285063,
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
        "name": "Josh Maint",
        "email": "wizzy-maint@wizzy.com"
      },
      "location": {
        "version": "bd34309759a381a850f60d90bb1adc2a1756bbdf11d746c438c8706a13c63f66",
        "path": "fake-leaks\\base64-encoded",
        "start": {
          "line": 2,
          "column": 2
        },
        "end": {
          "line": 10,
          "column": 13
        }
      },
      "notes": {}
    }
  ]
}
```

## TODO

1. Support local scans
1. Explore [base64 support](https://github.com/gitleaks/gitleaks/issues/807) (i.e. decoding it when spotted and scanning the contents)
1. Confirm full Mac support
1. Start integrating this into a pre-commit hooks project
1. Add `leaktk-scanner sanitize` to redact things from stdin for use in pipelines
1. Figure out a way to apply different rules in different contexts (internal/external repos, etc)
1. Support `.github/secret_scanning.yml` files
1. Have options for URL and JSONData to allow them to recursively pull other URLs when they see them (e.g. `follow_urls bool`, `depth uint16`) and make sure it can't loop
