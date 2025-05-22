# LeakTK

A growing toolkit of utilities for leak detection, mitigation and prevention.

## Usage Overview

```sh
# Run a single scan
leaktk scan https://github.com/leaktk/fake-leaks.git
leaktk scan --kind JSONData '{"key": "-----BEGIN PRIVATE KEY-----c5602d28d0f21422dfc7b572b17e6b138c1b49fd7f477d4c5c961e0756f1ff70-----END PRIVATE KEY-----"}'
leaktk scan --kind JSONData '@/path/to/some-file.json'

# Listen for requests
leaktk listen < ./examples/requests.jsonl

# See more options
leaktk help
```

When in `listen` mode, LeakTK listens on stdin, responds on stdout, and logs to
stderr. It reads one request per line and sends one response per line in
[JSONL](https://jsonlines.org/). It should always generate a response to each
request even if there were errors. More info on the formats are in the
[listen docs](./listen.md).

## Docs by Topic

- [Installation](./docs/install.md)
- [Scanning](./docs/scan.md)
- [Configuration](./docs/config.md)
- [Request/Response Formats for Listen](./listen.md)
- (Coming Soon) Monitoring Sources
- (Coming Soon) Analyzing Results
- (Coming Soon) Git Hook Setup
