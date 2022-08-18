# Scanner

Provides a consistent API around some existing scanning tools to integrate them with the rest of the tool kit.

## Status

Just getting started.

## Overview

The goal is to provide various input options (e.g. https/stdin) and various output options (webhook/stdout) that 
take a standardize [jsonl](https://jsonlines.org/) format that is independent of the underlying scanner.

The current targeted scanner is [gitleaks](https://github.com/zricethezav/gitleaks) because it is an awesome tool and we already have quite a few [patterns built up](https://github.com/leaktk/patterns) for it.

The point of this tool is to be able to either run a single instance in stand alone mode for easy scripting or run a cluster of them behind a loadbalencer as a part of a larger scanning pipeline.

## Scan Request Format

Schema WIP, but jsonl.

## Scan Results Format

Schema WIP, but jsonl.

## Config File Format

Schema WIP, but probably json.
