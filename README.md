# LSMV

## Overview

LSMV is a version control system inspired by git, but able to scale beyond a
single computer. Or at least it will be once it's finished.

## Why this project is necessary

The most common version control systems in use today are distributed version
control systems. They are distributed not in the sense of distributed systems
but in the sense that the entire repository is replicated on each developer's
machine instead of being centrally managed.

The problem with that, of course, is that when repositories get very large,
replicating them on every machine can become problematic or impossible. There is
a hard limit to the size a repository bound to a single computer can be, and
performance degrades well before that.

Large repositories can be useful for many reasons. For me, the most important
is monorepos. Monorepos dramatically simplify version control - you just keep
everything at head instead of pegging to specific versions. That also makes
developing libraries inside a large organization better, because you don't have
to worry about users getting out of date. If you need to push out a patch, you
just commit it and rebuild any binaries that consume the library.

But to enable performant monorepos at scale, we can't download the entire repo
onto every developer's computer. The goal of this project is to construct a
centralized version control system that's horizontally scalable for arbitrarily
large repositories and that allows developers to view and access the repo
without downloading the whole thing.

Because we're trying to escape the limitations of storing a repository on a
single computer, this version control system will be build on distributed
systems to eliminate having any single server as a point of failure or
bottleneck.

## Summary of components

### Objectstore

Objectstore is the service that provides long-term storage for all the commits,
files, and trees in a repository.

To add support for a new backend, you should implement the objectstore service
api (defined in `proto/objectstore/objectstore.proto`) in a new service instead
of extending an existing implementation.

For now all I've implemented is a simple in-memory store that works well enough
for testing but will not be useful in production.

### Versioning

The versioning server contains the main business logic for actually, like,
controlling versions. It provides the api that cli clients interact with for
most calls.

At the moment all it does is store and fetch commits. It will be extended with
more functionality as this project is developed.

### Fuse

Fuse is the key to the whole project. Fuse stands for Filesystems in USErspace,
and it's exactly what it sounds like: a filesystem defined in userspace. So by
implementing our own filesystem, we can make it appear that the entire
repository is present on disk while actually only fetching and caching the parts
of it we need at any given moment.

Additionally, we use fuse to expose a special control filesystem that lets users
and tools interact with the repository, along the lines of `/sys` and `/proc` on
linux. Ultimately we will provide a cli that uses these endpoints, but for now
I'm just manipulating the files directly to control the service.

## Building and developing this project

### Tools and requirements

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
author of this project doesn't care to develop on windows. For all they know, it
might work.

### Running the service

For now, you have to start all the servers manually. From the root of this
repository, run:
```
> bazel run //objectstore/bin:bin
```
Then in another terminal, run:
```
> bazel run //versioning/bin:bin
```
And finally, in a third terminal, run:
```
> bazel run //fuse:fuse -- $MOUNTLOCATION
```
With `$MOUNTLOCATION` replaced with the path to an existing, empty directory
where the filesystem will be mounted.

Then, you'll want to add some data to the server. In a fourth terminal:
```
> bazel run //dummy_data_pusher:dummy_data_pusher
```

Finally, you need to point the repository at the commit the dummy pusher just
loaded. Run:
```
echo 'fakecommit' > $MOUNTLOCATION/.control/head
```
Et viola! There should be a few files in `$MOUNTLOCATION/repo/`.
