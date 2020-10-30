---
title: "Similar projects"
linkTitle: "Similar projects"
weight: 90
description: >
  Similar projects.
---

- [vdeplug4](https://github.com/rd235/vdeplug4): vdeplug4 can create ad-hoc L2 networks over stdio.
  vdeplug4 is similar to NoRouter in the sense that it uses stdio, but vdeplug4 requires privileges (at least in userNS) for creating TAP devices.
  On ther other hand, NoRouter emulates L3 and does not need privileges, because NoRouter uses multi-loopback addresses such as 127.0.42.101 and 127.0.42.102.
  NoRouter also supports using HTTP and SOCKS proxies instead of multi-loopback addresses.

- [telepresence](https://www.telepresence.io/): Telepresence is kube-only and needs privileges.
  On ther other hand, NoRouter works with any environment and does not need privileges.

- `docker run -p` and `kubectl port-forward`: These port forwarders only support Local-to-Remote forwarding.
  On ther other hand, NoRouter supports omnidirectional forwarding: Local-to-Remote, Remote-to-Local, and Remote-to-Remote.
  And yet NoRouter is not specific to Docker/Kubernetes.

- `ssh -L` and `ssh -R`: `ssh -L` supports Local-to-Remote, `ssh -R` supports Remote-to-Local forwarding.
  On ther other hand, NoRouter supports omnidirectional forwarding: Local-to-Remote, Remote-to-Local, and Remote-to-Remote.
  And yet NoRouter is not specific to SSH.
