# Prometheus pushgateway client tool

[![Latest release](https://img.shields.io/github/v/release/hansmi/prompush)][releases]
[![Release workflow](https://github.com/hansmi/prompush/actions/workflows/release.yaml/badge.svg)](https://github.com/hansmi/prompush/actions/workflows/release.yaml)
[![CI workflow](https://github.com/hansmi/prompush/actions/workflows/ci.yaml/badge.svg)](https://github.com/hansmi/prompush/actions/workflows/ci.yaml)
[![Go reference](https://pkg.go.dev/badge/github.com/hansmi/prompush.svg)](https://pkg.go.dev/github.com/hansmi/prompush)

<!-- This repository hosts a Prometheus metrics exporter for -->
<!-- [Paperless-ngx][paperless], a document management system transforming physical -->
<!-- documents into a searchable online archive. The exporter relies on [Paperless' -->
<!-- REST API][paperless-api]. -->

<!-- An implementation using the API was chosen to provide the same perspective as -->
<!-- web browsers. -->


## Usage

`prompush` is a CLI program for pushing metrics to a [Prometheus
pushgateway](pushgateway).

See the `--help` output for usage.


## Installation

Pre-built binaries are provided for [all releases][releases].

Docker images via GitHub's container registry. The image supports Linux/AMD64
and Linux/ARM64.

```shell
docker pull ghcr.io/hansmi/prompush
```


[releases]: https://github.com/hansmi/prompush/releases/latest

<!-- vim: set sw=2 sts=2 et : -->
