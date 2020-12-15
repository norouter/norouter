---
title: "Download"
linkTitle: "Download"
weight: 2
description: >
  How to download and install NoRouter
---


The binary releases are available for Linux, macOS (Darwin), FreeBSD, NetBSD, OpenBSD, DragonFly BSD, and Windows.

Download from https://github.com/norouter/norouter/releases .

Or copy the following script to a terminal:

```bash
curl -fsSL https://github.com/norouter/norouter/releases/latest/download/norouter-$(uname -s)-$(uname -m).tgz | sudo tar xzvC /usr/local/bin
```

{{% alert %}}
**Note**

Make sure to use the (almost) same version of NoRouter across all the hosts.
{{% /alert %}}

{{% alert %}}
**Note**

The URL has changed in NoRouter v0.6.0.
{{% /alert %}}

## Already have a norouter binary?

When `norouter` is already installed on the local host, the `norouter show-installer` command can be used to replicate the
same version of the binary to remote hosts:

```console
$ norouter show-installer | ssh some-user@example.com
...
Successfully installed /home/some-user/bin/norouter (version 0.3.0)
```

