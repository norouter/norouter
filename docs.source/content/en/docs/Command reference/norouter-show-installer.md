---
title: "norouter show-installer"
linkTitle: "norouter show-installer"
weight: 41
---

`norouter show-install` the installer script to replicate the same version of `norouter` binary to other hosts.

The binary is located as `~/bin/norouter`.

## Examples

Show the installer script:

```console
$ norouter show-installer
#!/bin/sh
set -eu
# Installation script for NoRouter
# NOTE: make sure to use the same version across all the hosts.
...
```

Inject the script to a remote SSH host:

```console
$ norouter show-installer | ssh some-user@example.com
...
Successfully installed /home/some-user/bin/norouter (version 0.6.0)
```

## norouter show-installer --help
```
NAME:
   norouter show-installer - show script for installing NoRouter to other hosts

USAGE:
   norouter show-installer [command options] [arguments...]

OPTIONS:
   --version value  (default: "0.3.0")
   --help, -h       show help (default: false)
```
