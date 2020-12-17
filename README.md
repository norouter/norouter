![NoRouter banner](./docs.source/static/images/norouter_h.svg)

[NoRouter](https://norouter.io/) (IP-over-Stdio) is the easiest multi-host & multi-cloud networking ever:
- Works with any container, any VM, and any baremetal machine, on anywhere, as long as the shell access is available (e.g. `docker exec`, `kubectl exec`, `ssh`)
- Omnidirectional port forwarding: Local-to-Remote, Remote-to-Local, and Remote-to-Remote
- No routing configuration is required
- No root privilege is required (e.g. `sudo`, `docker run --privileged`)
- No public IP is required
- Provides several network modes
  - Loopback IP mode (e.g. 127.0.42.101, 127.0.42.102, ...)
  - HTTP proxy mode with built-in name resolver
  - SOCKS4a and SOCKS5 proxy mode with built-in name resolver
- Easily installable with a single binary, available for Linux, macOS, BSDs, and Windows

Web site: https://norouter.io/

- - -

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->


- [What is NoRouter?](#what-is-norouter)
- [Download](#download)
- [Quick usage](#quick-usage)
  - [Example 1: Port forwarding across localhost + Docker + Kubernetes + LXD + SSH](#example-1-port-forwarding-across-localhost--docker--kubernetes--lxd--ssh)
  - [Example 2: Virtual VPN connection into `docker network create` networks](#example-2-virtual-vpn-connection-into-docker-network-create-networks)
  - [Example 3: Virtual VPN connection into Kubernetes networks](#example-3-virtual-vpn-connection-into-kubernetes-networks)
  - [Example 4: Aggregate VPCs of AWS, Azure, and GCP](#example-4-aggregate-vpcs-of-aws-azure-and-gcp)
- [Documentation](#documentation)
- [Installing NoRouter from source](#installing-norouter-from-source)
- [Contributing to NoRouter](#contributing-to-norouter)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## What is NoRouter?

NoRouter implements unprivileged networking by using multiple loopback addresses such as 127.0.42.101 and 127.0.42.102.
The hosts in the network are connected by forwarding packets over stdio streams like `docker exec`, `kubectl exec`, `ssh`, and whatever.

Unlike traditional port forwarders such as `docker run -p`, `kubectl port-forward`, `ssh -L`, and `ssh -R`,
NoRouter provides mutual interconnectivity across multiple remote hosts.

![overview](./docs.source/static/images/norouter-overview.png)

NoRouter is mostly expected to be used in a dev environment for running heterogeneous multi-cloud apps.

e.g. An environment that is composed of:
- A laptop in the living room, for writing codes
- A baremetal workstation with GPU/FPGA in the office, for running machine-learning workloads
- ACI (Azure Container Instances) containers, for running other workloads that do not require a complete Kubernetes cluster
- EKS (Amazon Elastic Kubernetes Service) pods, for workloads that heavily access Amazon S3 buckets
- GKE (Google Kubernetes Engine) pods, for running gVisor-armored workloads

For production environments, setting up VPNs rather than NoRouter would be the right choice.

## Download

The binaries are available at https://github.com/norouter/norouter/releases .

See also [Getting Started](https://norouter.io/docs/getting-started/).

## Quick usage

- Install the `norouter` binary to all the hosts. Run `norouter show-installer` to show an installation script.
- Create a manifest YAML file. Run `norouter show-example` to show an example manifest.
- Run `norouter <FILE>` to start NoRouter with the specified manifest YAML file.

### Example 1: Port forwarding across localhost + Docker + Kubernetes + LXD + SSH

Run `norouter <FILE>` with the following YAML file:

```yaml
hosts:
# localhost
  local:
    vip: "127.0.42.100"
# Docker & Podman container (docker exec, podman exec)
  docker:
    cmd: "docker exec -i some-container norouter"
    vip: "127.0.42.101"
    ports: ["8080:127.0.0.1:80"]
# Writing /etc/hosts is possible on most Docker and Kubernetes containers
    writeEtcHosts: true
# Kubernetes Pod (kubectl exec)
  kube:
    cmd: "kubectl --context=some-context exec -i some-pod -- norouter"
    vip: "127.0.42.102"
    ports: ["8080:127.0.0.1:80"]
# Writing /etc/hosts is possible on most Docker and Kubernetes containers
    writeEtcHosts: true
# LXD container (lxc exec)
  lxd:
    cmd: "lxc exec some-container -- norouter"
    vip: "127.0.42.103"
    ports: ["8080:127.0.0.1:80"]
# SSH
# If your key has a passphrase, make sure to configure ssh-agent so that NoRouter can login to the remote host automatically.
  ssh:
    cmd: "ssh some-user@some-ssh-host.example.com -- norouter"
    vip: "127.0.42.104"
    ports: ["8080:127.0.0.1:80"]
```

In this example, 127.0.42.101:8080 on each hosts is forwarded to the port 80 of the Docker container.

Try:

```console
$ curl http://127.0.42.101:8080
$ docker exec some-container curl http://127.0.42.101:8080
$ kubectl --context=some-context exec some-pod -- curl http://127.0.42.101:8080
$ lxc exec some-container -- curl http://127.0.42.101:8080
$ ssh some-user@some-ssh-host.example.com -- curl http://127.0.42.101:8080
```

Similarly, 127.0.42.102:8080 is forwarded to the port 80 of the Kubernetes Pod,
127.0.42.103:8080 is forwarderd to the port 80 of the LXD container,
and 127.0.42.104:8080 is forwarded to the port 80 of `some-ssh-host.example.com`.

### Example 2: Virtual VPN connection into `docker network create` networks
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

Launch `norouter example2.yaml` with the following YAML:
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

### Example 3: Virtual VPN connection into Kubernetes networks

Example 2 can be also applied to Kubernetes clusters, just by replacing `docker exec` with `kubectl exec`.

```console
$ export http_proxy=http://127.0.0.1:18080
$ curl http://nginx.default.svc.cluster.local
```

### Example 4: Aggregate VPCs of AWS, Azure, and GCP

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

## Documentation

- [Top](https://norouter.io/docs/)
- [Getting Started](https://norouter.io/docs/getting-started/)
  - [Download](https://norouter.io/docs/getting-started/download/)
  - [First example](https://norouter.io/docs/getting-started/first-example/)
  - [Name resolution](https://norouter.io/docs/getting-started/name-resolution/)
  - [VPN-ish mode](https://norouter.io/docs/getting-started/vpn/)
- [Examples](https://norouter.io/docs/examples/)
  - [Docker](https://norouter.io/docs/examples/docker/)
  - [Podman](https://norouter.io/docs/examples/podman/)
  - [Kubernetes](https://norouter.io/docs/examples/kubernetes/)
  - [LXD](https://norouter.io/docs/examples/lxd/)
  - [SSH](https://norouter.io/docs/examples/ssh/)
  - [Azure Container Instances](https://norouter.io/docs/examples/azure-container-instances/)
  - [AWS/Azure/GCP VPCs](https://norouter.io/docs/examples/vpc/)
- [Troubleshooting](https://norouter.io/docs/troubleshooting/)
- [Command reference](https://norouter.io/docs/command-reference/)
  - [`norouter`](https://norouter.io/docs/command-reference/norouter/)
  - [`norouter manager`](https://norouter.io/docs/command-reference/norouter-manager/)
  - [`norouter agent`](https://norouter.io/docs/command-reference/norouter-agent/)
  - [`norouter show-example`](https://norouter.io/docs/command-reference/norouter-show-example/)
  - [`norouter show-installer`](https://norouter.io/docs/command-reference/norouter-show-installer/)
- [YAML reference](https://norouter.io/docs/yaml-reference/)
- [How it works](https://norouter.io/docs/how-it-works/)
- [Roadmap](https://norouter.io/docs/roadmap/)
- [Similar projects](https://norouter.io/docs/similar-projects/)
- [Artwork](https://norouter.io/docs/artwork/)

## Installing NoRouter from source

```console
$ make
$ sudo make install
```

## Contributing to NoRouter

- Please certify your [Developer Certificate of Origin (DCO)](https://developercertificate.org/), by signing off your commit with `git commit -s` and with your real name.
- Please add documents and tests whenever possible.

- - -

NoRouter is licensed under the terms of [Apache License, Version 2.0](./LICENSE).

Copyright (C) NoRouter authors.
