# LSMV

## Overview

LSMV is a version control system inspired by git, but able to scale beyond a single computer. Or at least it will be.

## Requirements

LSMV is built using bazel. We control versions with bazelisk, so you should
probably install and use that. If you don't want to use bazelisk, you can check
the `.bazelversion` file to see what version of bazel is expected to work.

Bazel will install a go toolchain and all necessary dependencies (including
gazelle) for you, so once you have bazel all targets in this repo should build
out-of-the-box.

Please generate your BUILD files with gazelle by running `bazel run //:gazelle`.
Also, please conform to gofmt style. Downloading and running gofmt is the
easiest way to do so.

Linux is the only supported platform for now. Macs cannot build certain targets
because fuse support was dropped for that platform in a dependency, and the
author of this repo doesn't care to develop on windows. For all they know, it
might work.
