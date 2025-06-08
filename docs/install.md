# Installation

There are several ways to install leaktk.

## **🐳 Using Docker/Podman (Container Image)**

Official container images are hosted on Quay.io.

```
# Replace TAG with a specific tag from https://quay.io/repository/leaktk/leaktk?tab=tags
podman run quay.io/leaktk/leaktk:TAG [command] [args...]
```

You can find available tags on [Quay.io for LeakTK](https://quay.io/repository/leaktk/leaktk?tab=tags).

## **💻 Pre-built Binaries**

Pre-built binaries for Linux, macOS, and Windows are available on the [GitHub
Releases page for leaktk/leaktk](https://github.com/leaktk/leaktk/releases).

1. Go to the [Releases page](https://github.com/leaktk/leaktk/releases).
2. Find the desired version.
3. Download the appropriate archive for your operating system and architecture
   (e.g., leaktk\_X.Y.Z\_linux\_amd64.tar.gz,
   leaktk\_X.Y.Z\_windows\_amd64.zip).
4. Extract the archive.
5. (Optional) Move the leaktk (or leaktk.exe on Windows) binary to a directory
   in your system's PATH (e.g., /usr/local/bin or C:\\Windows\\System32).

## **🛠️ Build From Source on Linux & macOS**

If you have Go installed (version 1.23.3 or newer is recommended), you can install leaktk directly using go install:

To install the latest version:

```sh
GOBIN="${HOME}/.local/bin" go install github.com/leaktk/leaktk@latest
```

Or to install a specific version:

```sh
# Replace vX.Y.Z with the specific tag from https://github.com/leaktk/leaktk/releases
GOBIN="${HOME}/.local/bin" go install github.com/leaktk/leaktk@vX.Y.X
```

You will want to make sure you have `"${HOME}/.local/bin"` in your `PATH` if it
isn't already.
