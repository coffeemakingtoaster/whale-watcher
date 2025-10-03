# Whale watcher

> This is a work in progress system! This tool as of now has a lot of quirks and servers as an unfinished proof-of-concept. Currently this is designed for macos and linux only. Windows support has never been tested.

Your way to define and enforce docker policies :) 

## Quick start

Install gopy:

```sh
make dep-install
```

Build (this also builds dependencies that are needed even if the program is started using go run);

```sh
make all
# Optional
sudo make install
```

Use tool: 
```sh
./build/whale-watcher help
```

The Dockerfile can be utilized for testing and the `_examples` directory contains sample rulesets.
To get an OCI tarball simply use `make oci-export`.
There is a [dummy repository](https://github.com/coffeemakingtoaster/whale-watcher-dummy-target) that can be tested against using ` WHALE_WATCHER_CONFIG_PATH=./_example/dummy_repository.config.yaml whale-watcher validate ./_example/big_example_ruleset.yaml`.

## Usage

Using whale watcher is rather straightforward. There are 2 possible modes of operation: `validate`, `docs`.

### Validate

Validate runs the validation of a given image based on a specified ruleset. The target image is retrieved based on the config, while the ruleset is passed in as a parameter.
```sh
whale-watcher validate <ruleset location>
```

This ruleset location may also be a git repository and a filepath within the repositor, specified in the format `<repo url ssh or http ending in .git>!<path>`

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
- [python3](https://www.python.org)
- [docker](https://docker.com)

## Troubleshooting

~~Building on modern MacOs  is broken due to LLVM versions not supporting old flags.~~
~~This requires altering and rebuilding gopy... See patched repo [here](https://github.com/coffeemakingtoaster/gopy).~~
