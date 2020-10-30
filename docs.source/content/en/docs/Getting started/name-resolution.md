---
title: "Name resolution"
linkTitle: "Name resolution"
weight: 4
description: >
  Using human-readable hostname strings instead of IP addresses
---

In the [First example](../first-example) we used IP addresses rather than hostnames because we cannot modify `/etc/hosts` without privileges.

For name resolution without privileges, NoRouter provides the following methods:
- Creating `$HOSTALIASES` file on each hosts
- Serving an HTTP proxy on each hosts
- Serving a SOCKS proxy on each hosts

However, when `/etc/hosts` is writable (mostly in Docker and Kubernetes), NoRouter can be also configured to write `/etc/hosts`,
by setting `.[]hosts.writeEtcHosts` to true.
See [Docker](../../examples/docker) and [Kubernetes](../../examples/kubernetes) examples.

## HOSTALIASES file
By default, NoRouter creates `~/.norouter/agent/hostaliases` file like this on each hosts:

```
host0 127.0.42.100.xip.io
host1 127.0.42.101.xip.io
host2 127.0.42.102.xip.io
```

The file can be used as `$HOSTALIASES` file if supported by applications.

```console
[localhost]$ HOSTALIASES=$HOME/.norouter/agent/hostaliases wget -O - http://host1:8080
[host1.cloud1.example.com]$ HOSTALIASES=$HOME/.norouter/agent/hostaliases wget -O - http://host1:8080
[host2.cloud2.example.com]$ HOSTALIASES=$HOME/.norouter/agent/hostaliases wget -O - http://host1:8080
```

Confirm that host1's Web service is shown.

{{% alert %}}
**Note**:

- Make sure to connect to 8080, not 80.
- Not all applications support resolving names with `$HOSTALIASES` file
- Hostnames with dots (e.g. "host1.norouter.local") is not added to `$HOSTALIASES` file
{{% /alert %}}


To change the directory for storing `$HOSTALIASES` file, set `.hostTemplate.stateDir.pathOnAgent` (or `.[]hosts.stateDir.pathOnAgent`) as follows:
```yaml
hostTemplate:
  stateDir:
    pathOnAgent: "~/foo/norouter-agent"
```

Creating `$HOSTALIASES` file is supported since NoRouter v0.4.0.

## HTTP proxy mode
To enable HTTP proxy mode, set `.hostTemplate.http.listen` (or `.[]hosts.http.listen`) as follows:

```yaml
hostTemplate:
  http:
    listen: "127.0.0.1:18080"
hosts:
  host0:
    vip: "127.0.42.100"
  host1:
    cmd: "ssh some-user@host1.cloud1.example.com -- /home/some-user/bin/norouter"
    vip: "127.0.42.101"
    ports: ["8080:127.0.0.1:80"]
  host2:
    cmd: "ssh some-user@host2.cloud2.example.com -- /home/some-user/bin/norouter"
    vip: "127.0.42.102"
    ports: ["8080:127.0.0.1:80"]
```

Applications needs `$http_proxy` and/or `$HTTP_PROXY` to be set to `http://127.0.0.1:18080`.

```console
[localhost]$ http_proxy=http://127.0.0.1:18080 wget -O - http://host1:8080
[host1.cloud1.example.com]$ http_proxy=http://127.0.0.1:18080 wget -O - http://host1:8080
[host2.cloud2.example.com]$ http_proxy=http://127.0.0.1:18080 wget -O - http://host1:8080
```

Confirm that host1's Web service is shown.

{{% alert %}}
**Note**:

Make sure to connect to 8080, not 80.
{{% /alert %}}

HTTP proxy mode is available since NoRouter v0.4.0.
### HTTP proxy mode without listening on multi-loopback addresses

When HTTP proxy mode is enabled, listening on multi-loopback addresses can be disabled by
setting `.hostTemplate.loopback.disable` (or `.[]hosts.loopback.disable`) to `true`.

e.g.

```yaml
hostTemplate:
  http:
    listen: "127.0.0.1:18080"
  loopback:
    disable: true
hosts:
  host0:
    vip: "127.0.42.100"
  host1:
    cmd: "ssh some-user@host1.cloud1.example.com -- /home/some-user/bin/norouter"
    vip: "127.0.42.101"
    ports: ["80:127.0.0.1:80"]
  host2:
    cmd: "ssh some-user@host2.cloud2.example.com -- /home/some-user/bin/norouter"
    vip: "127.0.42.102"
    ports: ["80:127.0.0.1:80"]
```

The manifest still contains virtual 127.0.42.{100,102,102} addresses, but these virtual
IP addresses are never listened by the real operating system.

Note that the port mapping is now changed from `8080:127.0.0.1:80` to `80:127.0.0.1:80`,
because we no longer need to avoid using privileged ports below 1024, and yet we no longer
need to care about port number collision .

Confirm that host1's Web service is shown with the following commands:
```console
[localhost]$ http_proxy=http://127.0.0.1:18080 wget -O - http://host1
[host1.cloud1.example.com]$ http_proxy=http://127.0.0.1:18080 wget -O - http://host1
[host2.cloud2.example.com]$ http_proxy=http://127.0.0.1:18080 wget -O - http://host1
```

## SOCKS proxy mode
In addition to HTTP proxy mode, NoRouter supports SOCKS proxy mode.

To enable SOCKS proxy mode, set `.hostTemplate.socks.listen` (or `.[]hosts.socks.listen`).

e.g.

```yaml
hostTemplate:
  http:
    listen: "127.0.0.1:18080"
  socks:
    listen: "127.0.0.1:18081"
...
```

NoRouter uses [github.com/cybozu-go/usockd/socks](https://pkg.go.dev/github.com/cybozu-go/usocksd/socks) for implementing SOCKS.
NoRouter supports SOCKS4, SOCKS4a, and SOCKS5.

SOCKS proxy mode is available since NoRouter v0.4.0.
