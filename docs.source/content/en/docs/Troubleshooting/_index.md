---
title: "Troubleshooting"
linkTitle: "Troubleshooting"
weight: 50
description: >
  Troubleshooting guide.
---

## Error "bind: can't assign requested address"
BSD hosts including macOS may face `listen tcp 127.0.43.101:8080: bind: can't assign requested address` error,
because BSDs only enable 127.0.0.1 as the loopback address by default.

A workaround is to run `sudo ifconfig lo0 alias 127.0.43.101`.

Solaris seems to require a similar workaround. (Help wanted.)

Also, gVisor is known to have a similar issue as of October 2020: https://github.com/google/gvisor/issues/4022
A workaround on gVisor is to run `ip addr add 127.0.0.2/8 dev lo`.

{{% alert %}}
**Note**

These multi-loopback addresses are not needed when you can use HTTP/SOCKS proxy mode.

e.g.
```yaml
hostTemplate:
  loopback:
# disables using multi-loopback addresses such as 127.0.43.101
    disable: true
  http:
# 127.0.0.1 can be always used
    listen: "127.0.0.1:18080"
  ports: ["80:127.0.0.1:80"]
hosts:
  host0: ...
  host1: ...
  host2: ...
```

```console
$ export http_proxy=127.0.0.1:18080
$ curl http://host1
$ curl http://host2
```
{{% /alert %}}
## Error "yaml: line N: found character that cannot start any token"

You might be mixing up tabs and spaces in the YAML.
