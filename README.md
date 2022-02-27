# LSMV

## Overview

LSMV is a version control system inspired by git, but able to scale beyond a single computer. Or at least it will be.

## Requirements

LSMV is built using bazel. We control versions with bazelisk, so you should probably install and use that.

Bazel will install a go toolchain for you if you don't already have one.

Please generate your BUILD files with gazelle by running `bazel run //:gazelle`.
