---
title: "Troubleshooting"
linkTitle: "Troubleshooting"
weight: 50
description: >
  Troubleshooting guide.
---

## Error "bind: can't assign requested address"
BSD hosts including macOS may face `listen tcp 127.0.43.101:8080: bind: can't assign requested address` error,
because BSDs only enable 127.0.0.1 as the loopback address by default.

A workaround is to run `sudo ifconfig lo0 alias 127.0.43.101`.

Solaris seems to require a similar workaround. (Help wanted.)

Also, gVisor is known to have a similar issue as of October 2020: https://github.com/google/gvisor/issues/4022
A workaround on gVisor is to run `ip addr add 127.0.0.2/8 dev lo`.
