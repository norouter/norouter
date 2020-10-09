
---
title: "NoRouter: IP-over-Stdio networking in a second"
linkTitle: "Documentation"
weight: 20
menu:
  main:
    weight: 20
---

NoRouter is the easiest multi-host & multi-cloud networking ever. And yet, NoRouter does not require any privilege such as `sudo` or `docker run --privileged`.

NoRouter implements unprivileged networking by using multiple loopback addresses such as 127.0.42.101 and 127.0.42.102.
The hosts in the network are connected by forwarding packets over stdio streams like `ssh`, `docker exec`, `podman exec`, `kubectl exec`, and whatever.

![overview](../images/norouter-overview.png)

{{% alert %}}
**Note**

NoRouter is mostly expected to be used in dev environments.
{{% /alert %}}
