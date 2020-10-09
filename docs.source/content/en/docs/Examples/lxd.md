---
title: "LXD"
linkTitle: "LXD"
weight: 20
---

Example manifest for LXD:

```yaml
hosts:
  lxd:
    cmd: "lxc exec some-container -- norouter"
    vip: "127.0.42.103"
    ports: ["8080:127.0.0.1:80"]
```

The `norouter` binary can be installed by using `lxc file push`:
```console
$ lxc launch ubuntu:20.04 foo
$ lxc file push norouter foo/usr/local/bin/norouter
```
