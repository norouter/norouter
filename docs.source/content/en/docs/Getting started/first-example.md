---
title: "First example"
linkTitle: "First example"
weight: 3
description: >
  The first example using two SSH hosts
---

Suppose that we have two remote hosts:

- `host1.cloud1.example.com`: running a Web service on TCP port 80
- `host2.cloud2.example.com`: running another Web service on TCP port 80

These hosts can be logged in from the local host via SSH.
However, these hosts are running on different clouds and they are NOT mutually IP-reachable.

The following example allows `host2` to connect to `host1` as `127.0.42.101:8080`,
and allows `host1` to connect to `host2` as `127.0.42.102:8080` using NoRouter.

## Step 0: Install `norouter

The `norouter` binary needs to be installed to all the remote hosts and the local host.
See [Download](../download).

The easiest way is to download the binary on the local host first, and then use
`norouter show-installer | ssh <USER>@<HOST>` to replicate the binary.

```console
$ curl -o norouter --fail -L https://github.com/norouter/norouter/releases/latest/download/norouter-$(uname -s)-$(uname -m)
$ chmod +x norouter
```

```console
$ norouter show-installer | ssh some-user@host1.cloud1.example.com
...
Successfully installed /home/some-user/bin/norouter (version 0.3.0)
```

```console
$ norouter show-installer | ssh some-user@host2.cloud2.example.com
...
Successfully installed /home/some-user/bin/norouter (version 0.3.0)
```

## Step 1: Create a manifest

Create a manifest file `norouter.yaml` on the local host as follows:

```yaml
hosts:
  # host0 is the localhost
  host0:
    vip: "127.0.42.100"
  host1:
    cmd: "ssh some-user@host1.cloud1.example.com -- /home/some-user/bin/norouter"
    vip: "127.0.42.101"
    ports: ["8080:127.0.0.1:80"]
  host2:
    cmd: "ssh some-user@host2.cloud2.example.com -- /home/some-user/bin/norouter"
    vip: "127.0.42.102"
    ports: ["8080:127.0.0.1:80"]
```

## Step 2: Start the main NoRouter process

```console
[localhost]$ ./bin/norouter norouter.yaml
```

If you are using macOS or BSD, you may see "bind: can't assign requested address" error.
See [Troubleshooting](../../troubleshooting) for a workaround.


## Step 3: connect to host1 (127.0.42.101)

```console
[localhost]$ wget -O - http://127.0.42.101:8080
[host1.cloud1.example.com]$ wget -O - http://127.0.42.101:8080
[host2.cloud2.example.com]$ wget -O - http://127.0.42.101:8080
```

Confirm that host1's Web service is shown.

{{% alert %}}
**Note**:

Make sure to connect to 8080, not 80.
{{% /alert %}}

## Step 4: connect to host2 (127.0.42.102)

```console
[localhost]$ wget -O - http://127.0.42.102:8080
[host1.cloud1.example.com]$ wget -O - http://127.0.42.102:8080
[host2.cloud2.example.com]$ wget -O - http://127.0.42.102:8080
```

Confirm that host2's Web service is shown.

