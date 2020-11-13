---
title: "VPN-ish mode"
linkTitle: "VPN-ish mode"
weight: 5
description: >
  Using NoRouter as a virtual VPN
---

Starting with NoRouter v0.5.0, NoRouter can be also used as a HTTP/SOCKS proxy that draws traffics
into a specific host.

This mode can be used like a virtual VPN.

## Virtual VPN connection into Docker networks

This example shows steps to use NoRouter for creating an HTTP proxy that works like a VPN router
that connects clients into `docker network create` networks.

This technique also works with remote Docker, rootless Docker, Docker for Mac, and even with Podman.
Read `docker` as `podman` for the usage with Podman.

First, create a Docker network named "foo", and create an nginx container named "nginx" there:
```console
$ docker network create foo
$ docker run -d --name nginx --hostname nginx --network foo nginx:alpine
```

Then, create a "bastion" container in the same network, and install NoRouter into it:
```console
$ docker run -d --name bastion --network foo alpine sleep infinity
$ norouter show-installer | docker exec -i bastion sh
```

Launch `norouter <FILE>` with the following YAML:
```yaml
hosts:
  local:
    vip: "127.0.42.100"
    http:
      listen: "127.0.0.1:18080"
    loopback:
      disable: true
  bastion:
    cmd: "docker exec -i bastion /root/bin/norouter"
    vip: "127.0.42.101"
routes:
  - via: bastion
    to: ["0.0.0.0/0", "*"]
```

The "nginx" container can be connected from the host as follows:
```console
$ export http_proxy=http://127.0.0.1:18080
$ curl http://nginx
```

If you are using Podman, try `curl http://nginx.dns.podman` rather than `curl http://nginx` .

## Virtual VPN connection into Kubernetes networks

The Docker example can be also applied to Kubernetes clusters, just by replacing `docker exec` with `kubectl exec`.

```console
$ export http_proxy=http://127.0.0.1:18080
$ curl http://nginx.default.svc.cluster.local
```

## Aggregate VPCs of AWS, Azure, and GCP

The following example provides an HTTP proxy that virtually aggregates VPCs of AWS, Azure, and GCP:

```yaml
hosts:
  local:
    vip: "127.0.42.100"
    http:
      listen: "127.0.0.1:18080"
  aws_bastion:
    cmd: "ssh aws_bastion -- ~/bin/norouter"
    vip: "127.0.42.101"
  azure_bastion:
    cmd: "ssh azure_bastion -- ~/bin/norouter"
    vip: "127.0.42.102"
  gcp_bastion:
    cmd: "ssh gcp_bastion -- ~/bin/norouter"
    vip: "127.0.42.103"
routes:
  - via: aws_bastion
    to:
      - "*.compute.internal"
  - via: azure_bastion
    to:
      - "*.internal.cloudapp.net"
  - via: gcp_bastion
    to:
# Substitute "example-123456" with your own GCP project ID
      - "*.example-123456.internal"
```

The localhost can access all remote hosts in these networks:
```console
$ export http_proxy=http://127.0.0.1:18080
$ curl http://ip-XXX-XXX-XX-XXX.ap-northeast-1.compute.internal
$ curl http://some-azure-host.internal.cloudapp.net
$ curl http://some-gcp-host.asia-northeast1-b.c.example-123456.internal
```

To allow accessing Azure and GCP networks from AWS hosts, set `.http.listen` of `aws_bastion` to `XXX.XXX.XXX.XXX:18080`, where `XXX.XXX.XXX.XXX` is a private IP of the AWS VPC.
Never use `0.0.0.0:18080` unless you have an appropriate firewall config:
