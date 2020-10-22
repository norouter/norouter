# NoRouter (IP-over-Stdio): the easiest multi-host & multi-cloud networking ever. No root privilege is required.

NoRouter is the easiest multi-host & multi-cloud networking ever. And yet, NoRouter does not require any privilege such as `sudo` or `docker run --privileged`.

Web site: https://norouter.io/

NoRouter implements unprivileged networking by using multiple loopback addresses such as 127.0.42.101 and 127.0.42.102.
The hosts in the network are connected by forwarding packets over stdio streams like `ssh`, `docker exec`, `podman exec`, `kubectl exec`, and whatever.

![overview](./docs.source/static/images/norouter-overview.png)

NoRouter is mostly expected to be used in dev environments.

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->


- [Download](#download)
- [Quick usage](#quick-usage)
- [Documentation](#documentation)
- [Installing NoRouter from source](#installing-norouter-from-source)
- [Contributing to NoRouter](#contributing-to-norouter)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Download

The binaries are available at https://github.com/norouter/norouter/releases .

See also [Getting Started](https://norouter.io/docs/getting-started/).

## Quick usage

- Install the `norouter` binary to all the hosts. Run `norouter show-installer` to show an installation script.
- Create a manifest YAML file. Run `norouter show-example` to show an example manifest.
- Run `norouter <FILE>` to start NoRouter with the specified manifest YAML file.

An example manifest file:
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
# Docker & Podman container (docker exec, podman exec)
# The cmd string can be also written as a string slice: ["docker", "exec", "-i", "some-container", "norouter"]
  docker:
    cmd: "docker exec -i some-container norouter"
    vip: "127.0.42.101"
    ports: ["8080:127.0.0.1:80"]
# Kubernetes Pod (kubectl exec)
  kube:
    cmd: "kubectl --context=some-context exec -i some-pod -- norouter"
    vip: "127.0.42.102"
    ports: ["8080:127.0.0.1:80"]
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

See [Documentation](#documentation) for the further information.

## Documentation

- [Top](https://norouter.io/docs/)
- [Getting Started](https://norouter.io/docs/getting-started/)
- [Examples](https://norouter.io/docs/examples/)
- [Troubleshooting](https://norouter.io/docs/troubleshooting/)
- [Command reference](https://norouter.io/docs/command-reference/)
- [YAML reference](https://norouter.io/docs/yaml-reference/)
- [How it works](https://norouter.io/docs/how-it-works/)
- [Roadmap](https://norouter.io/docs/roadmap/)
- [Similar projects](https://norouter.io/docs/similar-projects/)

## Installing NoRouter from source

```console
$ make
$ sudo make install
```

## Contributing to NoRouter

- Please sign-off your commit with `git commit -s` and with your real name.
- Please add documents and tests whenever possible.

- - -

NoRouter is licensed under the terms of [Apache License, Version 2.0](./LICENSE).

Copyright (C) NoRouter authors.
