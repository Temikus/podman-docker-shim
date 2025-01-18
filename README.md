# Podman-docker-shim

A simple shim for Podman to make it more compatible with Docker CLI.

## Supported commands

### push

  - `-a | --all-tags` - Push all tags of the image, see [podman issue #2369](https://github.com/containers/podman/issues/2369#issuecomment-1209431687)

## Building

```
docker build -t podman-docker-shim .
```
, or
```
mise  build
```

## Installation

```
cp podman-docker-shim /usr/local/bin
ln -s /usr/local/bin/podman-docker-shim /usr/local/bin/docker // or `mise setup`
```
, or
```
mise setup
```
