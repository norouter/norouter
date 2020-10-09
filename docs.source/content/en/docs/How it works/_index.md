---
title: "How NoRouter works under the hood"
linkTitle: "How it works"
weight: 70
description: >
  How NoRouter works under the hood.
---

The main NoRouter process launches the remote subprocesses and transfer L3 packets using their stdio streams.

To translate unprivileged socket syscalls into L3 packets, TCP/IP is implemented in userspace
using [netstack from gVisor & Fuchsia](https://pkg.go.dev/gvisor.dev/gvisor/pkg/tcpip/stack).

## Stdio packet protocol

```
uint8be  Magic     | 0x42
uint24be Len       | Length of the packet in bytes, excluding Magic and Len itself
uint16be Type      | 0x0001: L3, 0x0002: JSON (for configuration)
uint16be Reserved  | 0x0000
[]byte   L3OrJSON  | L3 or JSON
```

Further information:
* [`pkg/stream`](https://pkg.go.dev/github.com/norouter/norouter/pkg/stream)
* [`pkg/stream/jsonmsg`](https://pkg.go.dev/github.com/norouter/norouter/pkg/stream/jsonmsg)
