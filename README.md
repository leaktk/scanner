# Scanner

Provides a consistent API around some existing scanning tools to integrate them
with the rest of the tool kit.

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
leaktk-scanner /path/to/config.toml
```

The contents of the config file will determine how messages are sent/received,
where logs go, and other scanning behavior.

The "Scan Request Format" and "Scan Results Format" sections describe what you
should expect to see as input and output.

## Config File Format (WIP)

```toml
[logger]
level = "INFO"
filepath = "/var/log/leaktk/leaktk.log"

[listner]
# TODO: define method https settings (auth, tls, etc)
method = "stdin"

[scanner]
# This is the directory where the scanner will store files
workdir = "/tmp/leaktk"

[scanner.patterns]
# TODO: define auth settings for servers that require auth
server_url = "https://raw.githubusercontent.com/leaktk/patterns/main/target"
refresh_interval = 43200

[reporter] # This might get a better name soon. Still thinking on it.
# TODO: define method https settings (auth, tls, etc)
method = "stdout"
```

## Scan Request Format

Schema WIP, but likely jsonl.

## Scan Results Format

Schema WIP, but likely jsonl.
