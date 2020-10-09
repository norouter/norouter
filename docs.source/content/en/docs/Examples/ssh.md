---
title: "SSH"
linkTitle: "SSH"
weight: 30
---

Example manifest for remote SSH hosts:

```yaml
hosts:
  ssh:
    cmd: "ssh some-user@some-ssh-host.example.com -- norouter"
    vip: "127.0.42.104"
    ports: ["8080:127.0.0.1:80"]
```

If your key has a passphrase, make sure to configure `ssh-agent` so that NoRouter can login to the host automatically.

The `norouter` binary can be installed by using `scp`:
```console
$ scp cp norouter some-user@some-ssh-host.example.com:/usr/local/bin
```
