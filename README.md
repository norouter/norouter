# NoRouter: the easiest multi-host & multi-cloud networking ever. No root privilege is required.

NoRouter is the easiest multi-host & multi-cloud networking ever. And yet, NoRouter does not require any privilege such as `sudo` or `docker run --privileged`.

NoRouter implements unprivileged networking by using multiple loopback addresses such as 127.0.42.101 and 127.0.42.102.
The hosts in the network are connected by forwarding packets over stdio streams like `ssh`, `docker exec`, `podman exec`, `kubectl exec`, and whatever.

![./docs/image.png](./docs/image.png)


NoRouter is mostly expected to be used in dev environments.

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->


- [Download](#download)
- [Example using `docker exec` and `podman exec`](#example-using-docker-exec-and-podman-exec)
- [How it works under the hood](#how-it-works-under-the-hood)
  - [stdio protocol](#stdio-protocol)
- [More examples](#more-examples)
  - [Kubernetes](#kubernetes)
  - [SSH](#ssh)
  - [Azure Container Instances (`az container exec`)](#azure-container-instances-az-container-exec)
- [Troubleshooting](#troubleshooting)
  - [Error `bind: can't assign requested address`](#error-bind-cant-assign-requested-address)
- [TODOs](#todos)
- [Similar projects](#similar-projects)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Download

Download from https://github.com/norouter/norouter/releases .

To download using curl:
```
curl -o norouter --fail -L https://github.com/norouter/norouter/releases/latest/download/norouter-$(uname -s)-$(uname -m)
chmod +x norouter
```

> **Note**
>
> Make sure to use the (almost) same version of NoRouter across all the hosts.
> Notably, v0.1.x is completely incompatible with v0.2.x and newer versions.

## Example using `docker exec` and `podman exec`

This example creates a virtual 127.0.42.0/24 network across a Docker container, a Podman container, and the localhost, using `docker exec` and `podman exec`.

**Step 0: build `bin/norouter` binary** (on Linux)

```console
make
```

Or just download from [here](#Download).

**Step 1: create `host1` (nginx) as a Docker container**
```console
docker run -d --name host1 nginx:alpine
docker cp $(pwd)/bin/norouter host1:/usr/local/bin
```

**Step 2: create `host2` (Apache httpd) as a Podman container**
```console
podman run -d --name host2 httpd:alpine
podman cp $(pwd)/bin/norouter host2:/usr/local/bin
```

**Step 3: create [`example.yaml`](./example.yaml)**

```yaml
hosts:
  # host0 is the localhost
  host0:
    vip: "127.0.42.100"
  host1:
    cmd: ["docker", "exec", "-i", "host1", "norouter"]
    vip: "127.0.42.101"
    ports: ["8080:127.0.0.1:80"]
  host2:
    cmd: ["podman", "exec", "-i", "host2", "norouter"]
    vip: "127.0.42.102"
    ports: ["8080:127.0.0.1:80"]
```

**Step 4: start the main NoRouter process**

```console
./bin/norouter example.yaml
```

If you are using macOS or BSD, you may see "bind: can't assign requested address" error.
See [Troubleshooting](#troubleshooting) for a workaround.


**Step 5: connect to `host1` (127.0.42.101, nginx)**

```console
wget -O - http://127.0.42.101:8080
docker exec host1 wget -O - http://127.0.42.101:8080
podman exec host2 wget -O - http://127.0.42.101:8080
```

Confirm that nginx's `index.html` ("Welcome to nginx!") is shown.

> **Note**
>
> Make sure to connect to 8080, not 80.

**Step 6: connect to `host2` (127.0.42.102, Apache httpd)**

```console
wget -O - http://127.0.42.102:8080
docker exec host1 wget -O - http://127.0.42.102:8080
podman exec host2 wget -O - http://127.0.42.102:8080
```

Confirm that Apache httpd's `index.html` ("It works!") is shown.

## How it works under the hood

The main NoRouter process launches the following subprocesses and transfer L3 packets using their stdio streams.

* `/proc/self/exe internal agent` with configuration `Me="127.0.42.100"`, `Others={"127.0.42.101:8080", "127.0.42.102:8080"}`
* `docker exec -it host1 norouter internal agent` with `Me="127.0.42.101"`, `Forwards={"8080:127.0.0.1:80"}`, `Others={"127.0.42.102:8080"}`
* `docker exec -it host2 norouter internal agent` with `Me="127.0.42.102"`, `Forwards={"8080:127.0.0.1:80"}`, `Others={"127.0.42.101:8080"}`

`Me` is used as a virtual src IP for connecting to `Others`.

To translate unprivileged socket syscalls into L3 packets, TCP/IP is implemented in userspace
using [netstack from gVisor & Fuchsia](https://pkg.go.dev/gvisor.dev/gvisor/pkg/tcpip/stack).

### stdio protocol

This protocol is used since v0.2.0. Incompatible with v0.1.x.

```
uint8be  Magic     | 0x42
uint24be Len       | Length of the packet in bytes, excluding Magic and Len itself
uint16be Type      | 0x0001: L3, 0x0002: JSON (for configuration)
uint16be Reserved  | 0x0000
[]byte   L3OrJSON  | L3 or JSON
```

See [`pkg/stream`](./pkg/stream) for the further information.

## More examples

### Kubernetes

Install `norouter` binary using `kubectl cp`

e.g.
```
kubectl run --image=nginx:alpine --restart=Never nginx
kubectl cp bin/norouter nginx:/usr/local/bin
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

### SSH

Install `norouter` binary using `scp cp ./bin/norouter some-user@some-ssh-host.example.com:/usr/local/bin` .

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

## TODOs

- Assist generating mTLS certs?
- Add DNS fields to `/etc/resolv.conf` when the file is writable? (writable by default in Docker and Kubernetes)
- Detect port numbers automatically by watching `/proc/net/tcp`, and propagate the information across the cluster automatically?

## Similar projects

- [vdeplug4](https://github.com/rd235/vdeplug4): vdeplug4 can create ad-hoc L2 networks over stdio.
  vdeplug4 is similar to NoRouter in the sense that it uses stdio, but vdeplug4 requires privileges (at least in userNS) for creating TAP devices.
- [telepresence](https://www.telepresence.io/): kube-only and needs privileges

- - -

NoRouter is licensed under the terms of [Apache License, Version 2.0](./LICENSE).

Copyright (C) [Nippon Telegraph and Telephone Corporation](https://www.ntt.co.jp/index_e.html).
