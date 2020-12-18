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

## JSON messages

JSON messages are used to configure the agent. There are 3 types of messages:

- `request` are sent from the manager to an agent
- `result` messages are sent from an agent to the manager as a response to a `request`
- `event` are messages sent from an agent to the manager to indicate an independent event.

Messages always have the following structure:

```json
{
  "type": "request|response|event",
  "body": { ... }
}
```

### The `request` body

The request body has the following structure:

```json
{
    "id": 1, //Unique ID for this request
    "op": "operator", //Currently the only operator supported is "configure"
    "args": { ... }
}
```

### The `configure` message arguments

The `configure` message has the following arguments:

```json
{
  "me": "192.168.42.100",
  "forwards": [
    // See Forward below
  ],
  "others": [
    // See IPPortProto below
  ],
  "hostnameMap": {
     "hostname": "ip"
  },
  "http": {
    // See HTTP below
  },
  "socks": {
    // See "Socks" below
  },
  "loopback": {
    // See "Loopback" below
  },
  "statedir": {
    "path": "/home/user/.norouter/agent",
    "disable": false,
  },
  "writeEtcHosts": true,
  "routes": [
    // See "Route" below
  ],
  "nameServers": [
    // See IPPortProto below
  ]
}
```

### The `Forward` object

The forward object has the following fields:

```json
{
  "listen_port": 23
  // ConnectIP can be either IP or hostname.
  // Until NoRouter v0.4.0, it had to be IP.
  // The field name stil contains "_ip" suffix for compatibility.
  "connect_ip": "ip or hostname",
  "connect_port": 23,
  "proto": "tcp",
}
```

### The `IPPortProto` object

The `IPPortProto` object contains the following fields:

```json
{
  "ip": "192.168.42.100",
  "port": 23,
  "proto": "tcp"
}
```

### The `HTTP` object

```json
{
  "listen": "127.0.0.1:80"
}
```

### The `Socks` object

```json
{
  "listen": "127.0.0.1:5000"
}
```

### The `Loopback` object

```json
{
  "disable": true
}
``` 

### The `Route` object

```json
{
	"toCIDR": "192.168.95.0/24",
	"toHostnameGlob": "*.cloud1.example.com",
	"via": "192.168.42.100"
}
```

## The `result` body

```json
{
  "request_id": 1,
  "op": "configure",
  "error": {},
  "data": {}
}
```

### The `configure` result `data`

```json
{
    "features": [
      // Listening on multiple loopback IPs such as 127.0.42.101, 127.0.42.102, ...
      "loopback",
      // TCPv4 stream
      "tcp",
      // HTTP proxy
      "http",
      //Disable loopback
      "loopback.disable",
      //SOCKS
      "socks",
      //Creating ~/.norouter/agent/hostaliases file
      "hostaliases",
      //hostaliases using xip.io
      "hostaliases.\"xip.io\"",
      //Writing /etc/hosts when possible
      "etchosts",
      // Drawing packets into a specific hosts. Only meaningful for HTTP and SOCKS proxy modes
      "routes",
      //Built-in DNS
      "dns"
    ],
    "version": "norouter version"
}
```

## The `event` message

The event JSON looks like this:

```json
{
  "type": "event type",
  "data": { ... }
}
```

### The `routeSuggestion` event

This event suggests a route to the manager. The `data` field is as follows:

```json
{
  	"ip": [
      // List of IP addresses
    ],
  	"route": "192.168.42.100"
}
``` 

Further information:
* [`pkg/stream`](https://pkg.go.dev/github.com/norouter/norouter/pkg/stream)
* [`pkg/stream/jsonmsg`](https://pkg.go.dev/github.com/norouter/norouter/pkg/stream/jsonmsg)
