# Whale watcher

> This is a work in progress system! This tool as of now has a lot of quirks and servers as an unfinished proof-of-concept.

Your way to define and enforce docker policies :)

## Quick start

Install gopy:

```sh
make dep-install
```

Build (this also builds dependencies that are needed even if the program is started using go run);

```sh
make all
```

Use tool: 
```sh
./build/whale-watcher help
```

The Dockerfile can be utilized for testing and the `_examples` directory contains sample rulesets.
To get an OCI tarball simply use `make oci-export`.

## Usage

Using whale watcher is rather straightforward. There are 3 possible modes of operation: `bic`, `validate`, `docs`.

### Bic

Bic is short for base image cache which is used for building an index of the available base images.
Running `whale-watcher bci` will use all base images specified in the config and create a database of these base images and their components.

### Validate

Validate runs the validation of a given image based on a specified ruleset. The target image is retrieved based on the config, while the ruleset is passed in as a parameter.
```sh
whale-watcher validate <ruleset location>
```

This ruleset location may also be a git repository.

### Docs

Docs follows the same input format as validate, but rather than runnin validation logic it pretty prints the documentation for a given ruleset.

```sh
whale-watcher docs <ruleset location>
```

## Configuration

Configuring whale watcher can be done via the config file in YAML format (default location `./config.yaml`) and the file location can be specified using the `WHALE_WATCHER_CONFIG_PATH` environment variable.

For details on the configuration values see the [reference file](./reference.config.yaml).

All values can be overwritten using environment variables.
Generally these start with `WHALE_WATCHER_` and are followed by the keys of the yaml in capslock.
For instance the yaml field `github.pat` can be overwritten via `WHALE_WATCHER_GITHUB_PAT`.

## Development

Requirements:

- [gopy](https://github.com/go-python/gopy/tree/master)
- [python4](https://www.python.org)
- [docker](https://docker.com)

## Troubleshooting

Building on modern MacOs  is broken due to LLVM versions not supporting old flags.
This requires altering and rebuilding gopy...
