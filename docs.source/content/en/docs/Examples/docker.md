---
title: "Docker"
linkTitle: "Docker"
weight: 10
---

Example manifest for Docker:

```yaml
hosts:
  docker:
    cmd: "docker exec -i some-container norouter"
    vip: "127.0.42.101"
    ports: ["8080:127.0.0.1:80"]
# Writing /etc/hosts is possible on most Docker and Kubernetes containers
    writeEtcHosts: true
```

The `norouter` binary can be installed by using `docker cp`:
```console
$ docker run -d --name foo nginx:alpine
$ docker cp norouter foo:/usr/local/bin
```

## Virtual VPN connection into Docker networks

NoRouter also supports creating an HTTP proxy that works like a VPN router that connects clients into `docker network create` networks.

This technique also works with remote Docker, rootless Docker, Docker for Mac, and even with Podman.

See [Getting Started/VPN-ish mode](../../getting-started/vpn).
