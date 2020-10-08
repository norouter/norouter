# NoRouter: the easiest multi-host & multi-cloud networking ever. No root privilege is required.

NoRouter is the easiest multi-host & multi-cloud networking ever. And yet, NoRouter does not require any privilege such as `sudo` or `docker run --privileged`.

NoRouter implements unprivileged networking by using multiple loopback addresses such as 127.0.42.101 and 127.0.42.102.
The hosts in the network are connected by forwarding packets over stdio streams like `ssh`, `docker exec`, `podman exec`, `kubectl exec`, and whatever.

![./docs/image.png](./docs/image.png)


NoRouter is mostly expected to be used in dev environments.

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->


- [Download](#download)
- [Example using two SSH hosts](#example-using-two-ssh-hosts)
- [How it works under the hood](#how-it-works-under-the-hood)
  - [stdio protocol](#stdio-protocol)
- [More examples](#more-examples)
  - [Docker](#docker)
  - [Podman](#podman)
  - [Kubernetes](#kubernetes)
  - [LXD](#lxd)
  - [SSH](#ssh)
  - [Azure Container Instances (`az container exec`)](#azure-container-instances-az-container-exec)
- [Troubleshooting](#troubleshooting)
  - [Error `bind: can't assign requested address`](#error-bind-cant-assign-requested-address)
- [TODOs](#todos)
- [Compile NoRouter from the source](#compile-norouter-from-the-source)
- [Contributing to NoRouter](#contributing-to-norouter)
- [Similar projects](#similar-projects)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Download

The binary releases are available for Linux, macOS (Darwin), FreeBSD, NetBSD, OpenBSD, DragonFly BSD, and Windows.

Download from https://github.com/norouter/norouter/releases .

To download using curl:
```bash
curl -o norouter --fail -L https://github.com/norouter/norouter/releases/latest/download/norouter-$(uname -s)-$(uname -m)
chmod +x norouter
```

> **Note**
>
> Make sure to use the (almost) same version of NoRouter across all the hosts.
> Notably, v0.1.x is completely incompatible with v0.2.x and newer versions.

When `norouter` is already installed on the local host, the `norouter show-installer` command can be used to replicate the
same version of the binary to remote hosts:

```console
$ norouter show-installer | ssh some-user@example.com
...
Successfully installed /home/some-user/bin/norouter (version 0.2.0)
```

## Example using two SSH hosts

Suppose that we have two remote hosts:

- `host1.cloud1.example.com`: running a Web service on TCP port 80
- `host2.cloud2.example.com`: running another Web service on TCP port 80

These hosts can be logged in from the local host via SSH.
However, these hosts are running on different clouds and they are NOT mutually IP-reachable.

The following example allows `host2` to connect to `host1` as `127.0.42.101:8080`,
and allows `host1` to connect to `host2` as `127.0.42.102:8080` using NoRouter.

**Step 0: Install `norouter`**

The `norouter` binary needs to be installed to all the remote hosts and the local host.
See [Download](#download).

```console
[localhost]$ curl -o norouter --fail -L https://github.com/norouter/norouter/releases/latest/download/norouter-$(uname -s)-$(uname -m)

[localhost]$ chmod +x norouter

[localhost]$ norouter show-installer | ssh some-user@host1.cloud1.example.com
...
Successfully installed /home/some-user/bin/norouter (version 0.2.0)

[localhost]$ norouter show-installer | ssh some-user@host2.cloud2.example.com
...
Successfully installed /home/some-user/bin/norouter (version 0.2.0)
```

**Step 1: create a manifest**

Create a manifest file `norouter.yaml` on the local host as follows:

```yaml
hosts:
  # host0 is the localhost
  host0:
    vip: "127.0.42.100"
  host1:
    cmd: ["ssh", "some-user@host1.cloud1.example.com", "--", "/home/some-user/bin/norouter"]
    vip: "127.0.42.101"
    ports: ["8080:127.0.0.1:80"]
  host2:
    cmd: ["ssh", "some-user@host2.cloud2.example.com", "--", "/home/some-user/bin/norouter"]
    vip: "127.0.42.102"
    ports: ["8080:127.0.0.1:80"]
```

**Step 2: start the main NoRouter process**

```console
[localhost]$ ./bin/norouter norouter.yaml
```

If you are using macOS or BSD, you may see "bind: can't assign requested address" error.
See [Troubleshooting](#troubleshooting) for a workaround.


**Step 3: connect to `host1` (127.0.42.101)**

```console
[localhost]$ wget -O - http://127.0.42.101:8080
[host1.cloud1.example.com]$ wget -O - http://127.0.42.101:8080
[host2.cloud2.example.com]$ wget -O - http://127.0.42.101:8080
```

Confirm that host1's Web service is shown.

> **Note**
>
> Make sure to connect to 8080, not 80.

**Step 4: connect to `host2` (127.0.42.102)**

```console
[localhost]$ wget -O - http://127.0.42.102:8080
[host1.cloud1.example.com]$ wget -O - http://127.0.42.102:8080
[host2.cloud2.example.com]$ wget -O - http://127.0.42.102:8080
```

Confirm that host2's Web service is shown.

## How it works under the hood

The main NoRouter process launches the remote subprocesses and transfer L3 packets using their stdio streams.

To translate unprivileged socket syscalls into L3 packets, TCP/IP is implemented in userspace
using [netstack from gVisor & Fuchsia](https://pkg.go.dev/gvisor.dev/gvisor/pkg/tcpip/stack).

### stdio protocol

This protocol is used since NoRouter v0.2.0. Incompatible with v0.1.x.

```
uint8be  Magic     | 0x42
uint24be Len       | Length of the packet in bytes, excluding Magic and Len itself
uint16be Type      | 0x0001: L3, 0x0002: JSON (for configuration)
uint16be Reserved  | 0x0000
[]byte   L3OrJSON  | L3 or JSON
```

See [`pkg/stream`](./pkg/stream) for the further information.

## More examples

See [`example.yaml`](./example.yaml):
```yaml
# Example manifest for NoRouter.
# Run `norouter <FILE>` to start NoRouter with the specified manifest file.
#
# The `norouter` binary needs to be installed on all the remote hosts.
# Run `norouter show-installer` to show the installation script.
#
hosts:
# localhost
  local:
    vip: "127.0.42.100"
# Docker container (docker exec)
  docker:
    cmd: ["docker", "exec", "-i", "some-container", "norouter"]
    vip: "127.0.42.101"
    ports: ["8080:127.0.0.1:80"]
# Podman container (podman exec)
  podman:
    cmd: ["podman", "exec", "-i", "some-container", "norouter"]
    vip: "127.0.42.102"
    ports: ["8080:127.0.0.1:80"]
# Kubernetes Pod (kubectl exec)
  kube:
    cmd: ["kubectl", "--context=some-context", "exec", "-i", "some-pod", "--", "norouter"]
    vip: "127.0.42.103"
    ports: ["8080:127.0.0.1:80"]
# SSH
# If your key has a passphrase, make sure to configure ssh-agent so that NoRouter can login to the remote host automatically.
  ssh:
    cmd: ["ssh", "some-user@some-ssh-host.example.com", "--", "norouter"]
    vip: "127.0.42.104"
    ports: ["8080:127.0.0.1:80"]
```

The example can be also shown by running `norouter show-example`, or by running `norouter --open-editor`.

### Docker

Install `norouter` binary using `docker cp`:
```
docker run -d --name foo nginx:alpine
docker cp norouter foo:/usr/local/bin
```

In the NoRouter yaml, specify `cmd` as `["docker", "exec", "-i", "foo", "norouter"]`.

### Podman

Same as [Docker](#docker), but read `docker` as `podman`.

### Kubernetes

Install `norouter` binary using `kubectl cp`:
```
kubectl run --image=nginx:alpine --restart=Never nginx
kubectl cp norouter nginx:/usr/local/bin
```

In the NoRouter yaml, specify `cmd` as `["kubectl", "exec", "-i", "some-kubernetes-pod", "--", "norouter"]`.
To connect multiple Kubernetes clusters, pass `--context` arguments to `kubectl`.

e.g. To connect GKE, AKS, and your laptop:

```yaml
hosts:
  laptop:
    vip: "127.0.42.100"
  nginx-on-gke:
    cmd: ["kubectl", "--context=gke_myproject-12345_asia-northeast1-c_my-gke", "exec", "-i", "nginx", "--", "norouter"]
    vip: "127.0.42.101"
    ports: ["8080:127.0.0.1:80"]
  httpd-on-aks:
    cmd: ["kubectl", "--context=my-aks", "exec", "-i", "httpd", "--", "norouter"]
    vip: "127.0.42.102"
    ports: ["8080:127.0.0.1:80"]
```

### LXD

Install `norouter` binary using `lxc file push`:

```
lxc launch ubuntu:20.04 foo
lxc file push norouter foo/usr/local/bin/norouter
```

In the NoRouter yaml, specify `cmd` as `["lxc", "exec", "foo", "--", norouter"]`.

### SSH

Install `norouter` binary using `scp cp norouter some-user@some-ssh-host.example.com:/usr/local/bin` .

In the NoRouter yaml, specify `cmd` as `["ssh", "some-user@some-ssh-host.example.com", "--", "norouter"]`.

If your key has a passphrase, make sure to configure `ssh-agent` so that NoRouter can login to the host automatically.

### Azure Container Instances (`az container exec`)

`az container exec` can't be supported currently because:
- No support for stdin without tty: https://github.com/Azure/azure-cli/issues/15225
- No support for appending command arguments: https://docs.microsoft.com/en-us/azure/container-instances/container-instances-exec#restrictions
- Extra TTY escape sequence on busybox: https://github.com/Azure/azure-cli/issues/6537

A workaround is to inject an SSH sidecar into an Azure container group, and use `ssh` instead of `az container exec`.

## Troubleshooting

### Error `bind: can't assign requested address`
BSD hosts including macOS may face `listen tcp 127.0.43.101:8080: bind: can't assign requested address` error,
because BSDs only enable 127.0.0.1 as the loopback address by default.

A workaround is to run `sudo ifconfig lo0 alias 127.0.43.101`.

Solaris seems to require a similar workaround. (Help wanted.)

Also, gVisor is known to have a similar issue as of October 2020: https://github.com/google/gvisor/issues/4022
A workaround on gVisor is to run `ip addr add 127.0.0.2/8 dev lo`.

## TODOs

- Assist generating mTLS certs?
- Add DNS fields to `/etc/resolv.conf` when the file is writable? (writable by default in Docker and Kubernetes)
- Generate [`HOSTALIASES` file](https://man7.org/linux/man-pages/man7/hostname.7.html)?
- Create Kubernetes `Services`?
- Detect port numbers automatically by watching `/proc/net/tcp`, and propagate the information across the cluster automatically?

## Compile NoRouter from the source

Just run `make`.

## Contributing to NoRouter

- Please sign-off your commit with `git commit -s` and with your real name.
- Please add documents and tests whenever possible.

## Similar projects

- [vdeplug4](https://github.com/rd235/vdeplug4): vdeplug4 can create ad-hoc L2 networks over stdio.
  vdeplug4 is similar to NoRouter in the sense that it uses stdio, but vdeplug4 requires privileges (at least in userNS) for creating TAP devices.
- [telepresence](https://www.telepresence.io/): kube-only and needs privileges

- - -

NoRouter is licensed under the terms of [Apache License, Version 2.0](./LICENSE).

Copyright (C) NoRouter authors.
