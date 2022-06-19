# carousel

the iac wrapper that will continue going in a circle to deploy servers.

[![Build Status](https://github.com/xmidt-org/carousel/actions/workflows/ci.yml/badge.svg)](https://github.com/xmidt-org/carousel/actions/workflows/ci.yml)
[![Dependency Updateer](https://github.com/xmidt-org/carousel/actions/workflows/updater.yml/badge.svg)](https://github.com/xmidt-org/carousel/actions/workflows/updater.yml)
[![codecov.io](http://codecov.io/github/xmidt-org/carousel/coverage.svg?branch=main)](http://codecov.io/github/xmidt-org/carousel?branch=main)
[![Go Report Card](https://goreportcard.com/badge/github.com/xmidt-org/carousel)](https://goreportcard.com/report/github.com/xmidt-org/carousel)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=xmidt-org_carousel&metric=alert_status)](https://sonarcloud.io/dashboard?id=xmidt-org_carousel)
[![Apache V2 License](http://img.shields.io/badge/license-Apache%20V2-blue.svg)](https://github.com/xmidt-org/carousel/blob/main/LICENSE)
[![GitHub Release](https://img.shields.io/github/release/xmidt-org/carousel.svg)](CHANGELOG.md)

## Summary

Carousel is the micromanager of an IaC to help manage errors even those that are not detected.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Details](#details)
- [Install](#install)
- [Contributing](#contributing)

## Code of Conduct

This project and everyone participating in it are governed by
the [XMiDT Code Of Conduct](https://xmidt.io/code_of_conduct/). By participating, you agree to this Code.

## Details

## OUT OF SCOPE

carousel will NOT set up the environment required for the IaC. In other words, terraform must be able to run
successfully before using carousel.

## Terraform

### PreReq

The `.tf` file MUST contain the following:

```hcl
variable "versionBlueCount" {
  description = "Number of instances for version Blue"
  default     = 0
}

variable "versionBlue" {
  description = "version for the software of group Blue"
  default     = "0.0.0"
}

variable "versionGreenCount" {
  description = "Number of instances for version Green"
  default     = 0
}

variable "versionGreen" {
  description = "version for the software of group Green"
  default     = "0.0.0"
}

module "green" {
  ...
}
module "blue" {
  ...
}

output "blueHostnames" {
  value = module.blue.fqdn
}
output "greenHostnames" {
  value = module.green.fqdn
}

output "blueVersion" {
  value = var.versionBlue
}
output "greenVersion" {
  value = var.versionGreen
}
```

### Simple Run

```bash
# For a dry run of upgrading to 0.3.1 with 4 nodes
carousel rollout -d 4 0.3.1
```

### Host Validation

It is possible to provide a [golang plugin](https://golang.org/pkg/plugin/) to check a created host. Build a golang
plugins with the Func `func CheckHost(fqdn string) bool` defined.

```bash
# For rollout 4 nodes of version 1.2.3
# the created hosts will be validated against the CheckHost(fqdn string) bool func
carousel rollout -p hostValidator.so 4 1.2.3
```

For more information refer to the [example dir](./example/README.md)

## Docker

```bash
make docker
# note plugins must be compiled on the same OS.

docker run --rm -v carousel.yaml:/carousel.yaml -v deployment:/deployment/ -e WORK_DIR=/deployment/ carousel:latest
```

## Build

In order to build from the source, you need a working Go environment with version 1.16 or greater. Find more information
on the [Go website](https://golang.org/doc/install).

You can directly use `go get` to put the carousel binary into your `GOPATH`:

```bash
go get github.com/xmidt-org/carousel
```

You can also clone the repository yourself and build using make:

```bash
mkdir -p $GOPATH/src/github.com/xmidt-org/carousel
cd $GOPATH/src/github.com/xmidt-org/carousel
git clone git@github.com:xmidt-org/carousel.git
cd carousel
make build
```

## Contributing

Refer to [CONTRIBUTING.md](CONTRIBUTING.md).

## Credits

carousel would not be possible without the help of many other pieces of open source software. Thank you open source
world!

Given the Apache 2.0 license of carousel, we specifically want to call out the following libraries and their
corresponding licenses shown below.

- [terraform](https://github.com/hashicorp/terraform) - [MPL-2.0](https://www.mozilla.org/MPL/2.0/)
