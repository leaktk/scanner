# Scanner

Provides a consistent API around some existing scanning tools to integrate them with the rest of the tool kit.

## Status

Just getting started.

## Overview

This scanner is meant to either be ran as a single instance listening on stdin
for easy scripting or ran as a cluster behind a HTTP load balencer as a part of
a larger scanning pipeline.

The scanner leverages
[gitleaks](https://github.com/zricethezav/gitleaks)
internally because gitleaks is an awesome tool, and we already have quite a few
[patterns built up](https://github.com/leaktk/patterns)
for it.

## Usage (Pending Implementation)

To start the scanner and have it listen for requests run:

```sh
leaktk-scanner /path/to/config.json
```

The contents of the config file will determine how messages are sent/received,
where logs go, and other scanning behavior.

The "Scan Request Format" and "Scan Results Format" sections describe what you
should expect to see as input and output.

## Config File Format

The config file should contain a JSON object with the following attrs:

* `logs` - (optional) config for logging
* `recv` - config for receiving requests
* `scan` - config for scanning
* `send` - config for sending results

Format:

```
{
  "logs": <logs-config-object>,
  "recv": <recv-config-object>,
  "scan": <scan-config-object>,
  "send": <send-config-object>
}
```

The sub-sections below provide more info about the available options.

### logs-config-object

The logs-config-object should contain a JSON object with the following attrs:

* `level` - (optional) one of `DEBUG`, `INFO`, `WARN`, `ERROR`, `CRIT` (defaults to `INFO`)
* `filepath` - (optional) where to write the logs (defaults to stderr)

Example:

```
{
  "level": "INFO",
  "filepath": "/var/log/leaktk/leaktk.log"
}
```

### recv-config-object

The recv-config-object should contain a JSON object with the following attrs:

* `method` - defines how it should listen (options: `stdin`, `https`)
* `auth` - (only for method=https) an auth-config-object
* `tls` - (only for method=https) a tls-config-object

Example:

```
{
  "method": "stdin"
}
```

### send-config-object

The send-config-object should contain a JSON object with the following attrs:

* `method` - defines where it should send results (options: `stdout`, `https`)
* `auth` - (optional and only for method=https) an auth-config-object
* `tls` - (optional and only for method=https) a tls-config-object

Example:

```
{
  "method": "stdout"
}
```

### scan-config-object

The scan-config-object should contain a JSON object with the following attrs:

* `patterns` - a patterns-config-object for pulling patterns
* `workdirpath` - a directory for the scanner to put files in
* `defaults` - default config options by scan request type

Example:

```
{
  "patterns": <patterns-config-object>,
  "workdirpath": "/tmp/leaktk",
  "defaults": {
    "git": {
      "clone.depth": "250",
    }
  }
}
```

### patterns-config-object

The patterns-config-object should contain a JSON object with the following attrs:

* `serverurl` - the base pattern server url without `/patterns/...`
* `interval` - how many seconds to wait between updates
* `auth` - (optional) a auth-config-object
* `tls` -  (optional) a tls-config-object

Example:

```
{
  "serverurl": "https://raw.githubusercontent.com/leaktk/patterns/main/target",
  "interval": 43200
}
```

### auth-config-object

The auth-config-object should contain a JSON object with the following attrs:

(TBD)

### tls-config-object

The tls-config-object should contain a JSON object with the following attrs:

(TBD)

## Scan Request Format

Schema WIP, but jsonl.

## Scan Results Format

Schema WIP, but jsonl.
