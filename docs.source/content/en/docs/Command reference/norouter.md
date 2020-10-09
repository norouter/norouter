---
title: "norouter"
linkTitle: "norouter"
weight: 10
---

The `norouter` command is the top-level command of NoRouter.

When the first argument is a file name, `norouter` automatically executes [`norouter manager <FILE>`](../norouter-manager/).
Otherwise the argument is assumed to be a subcommand.

## Examples

Run Norouter using a manifest file `example.yaml`:

```console
$ norouter example.yaml
```

Same as the first example but explicitly specify [the `manager` subcommand](../norouter-manager/):

```console
$ norouter manager example.yaml
```

### --open-editor (-e)

Open an editor for a temporary manifest file, with an example content:

```console
$ norouter -e
```

### --version (-v)

Show the NoRouter version:

```console
$ norouter -v
norouter version 0.3.0
```

## norouter --help

```
NAME:
   norouter - the easiest multi-host & multi-cloud networking ever. No root privilege is required.

USAGE:
   norouter [global options] command [command options] [arguments...]

VERSION:
   0.3.0

DESCRIPTION:
   
  NoRouter is the easiest multi-host & multi-cloud networking ever.
  And yet, NoRouter does not require any privilege such as `sudo` or `docker run --privileged`.

  NoRouter implements unprivileged networking by using multiple loopback addresses such as 127.0.42.101 and 127.0.42.102.
  The hosts in the network are connected by forwarding packets over stdio streams like `ssh`, `docker exec`, `podman exec`, `kubectl exec`, and whatever.

  Quick usage:
  - Install the `norouter` binary to all the hosts. Run `norouter show-installer` to show an installation script.
  - Create a manifest YAML file. Run `norouter show-example` to show an example manifest.
  - Run `norouter <FILE>` to start NoRouter with the specified manifest YAML file.

  Documentation: https://github.com/norouter/norouter

COMMANDS:
   manager, m             manager (default subcommand)
   agent                  agent (No need to launch manually)
   show-example, show-ex  show an example manifest
   show-installer         show script for installing NoRouter to other hosts
   help, h                Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --debug            debug mode (default: false)
   --open-editor, -e  open an editor for a temporary manifest file, with an example content (default: false)
   --help, -h         show help (default: false)
   --version, -v      print the version (default: false)
```
