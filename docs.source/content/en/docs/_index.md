
---
title: "NoRouter (IP-over-Stdio): Unprivileged instant multi-cloud networking"
linkTitle: "Documentation"
weight: 20
menu:
  main:
    weight: 20
---

NoRouter is the easiest multi-host & multi-cloud networking ever:

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


NoRouter implements unprivileged networking by using multiple loopback addresses such as 127.0.42.101 and 127.0.42.102.

The hosts in the network are connected by forwarding packets over stdio streams like `docker exec`, `kubectl exec`, `ssh`, and whatever.

Unlike traditional port forwarders such as `docker run -p`, `kubectl port-forward`, `ssh -L`, and `ssh -R`,
NoRouter provides mutual interconnectivity across multiple remote hosts.

![overview](../images/norouter-overview.png)

{{% alert %}}
**Note**

NoRouter is mostly expected to be used in a dev environment for running heterogeneous multi-cloud apps.

e.g. An environment that is composed of:
- A laptop in the living room, for writing codes
- A baremetal workstation with GPU/FPGA in the office, for running machine-learning workloads
- ACI (Azure Container Instances) containers, for running other workloads that do not require a complete Kubernetes cluster
- EKS (Amazon Elastic Kubernetes Service) pods, for workloads that heavily access Amazon S3 buckets
- GKE (Google Kubernetes Engine) pods, for running gVisor-armored workloads

For production environments, setting up VPNs rather than NoRouter would be the right choice.

{{% /alert %}}
