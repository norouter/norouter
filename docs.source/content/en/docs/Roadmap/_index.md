---
title: "Roadmap"
linkTitle: "Roadmap"
weight: 85
description: >
  Project roadmap.
---

(To be determined)

- Assist generating mTLS certs?
- Launch a virtual DNS when `CAP_NET_BIND_SERVICE` is granted, and add the DNS to `/etc/resolv.conf` when the file is writable? (writable by default in Docker and Kubernetes)
- Generate [`HOSTALIASES` file](https://man7.org/linux/man-pages/man7/hostname.7.html)?
- Create Kubernetes `Services`?
- Detect port numbers automatically by watching `/proc/net/tcp`, and propagate the information across the cluster automatically?

- - -

See also https://github.com/norouter/norouter/issues
