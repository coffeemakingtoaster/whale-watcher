# Whale watcher

Your way to define and enforce docker policies :)

## Quick start

Install gopy:

```sh
make dep-install
```

Build:

```sh
make all
```

Use tool: 
```sh
./build/whale-watcher help
```

The Dockerfile can be utilized for testing and the `_examples` directory contains sample rulesets.
To get an OCI tarball simply use `make oci-export`.

## Development

Requirements:

- [gopy](https://github.com/go-python/gopy/tree/master)
- [python3](https://www.python.org)

## Troubleshooting

Building on modern MacOs  is broken due to LLVM versions not supporting old flags.
This requires altering and rebuilding gopy...
