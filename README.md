# NoRouter: the easiest multi-host & multi-cloud networking ever. No root privilege required.

NoRouter is the easiest multi-host & multi-cloud networking ever. And yet, NoRouter does not require any privilege such as `sudo` or `docker run --privileged`.

NoRouter implements unprivileged networking by using multiple loopback addresses such as 127.0.42.101 and 127.0.42.102.
The hosts in the network are connected by forwarding packets over stdio streams like `ssh`, `docker exec`, `podman exec`, `kubectl exec`, and whatever.

![./docs/image.png](./docs/image.png)


NoRouter is mostly expected to be used in dev environments.

## Example using `docker exec` and `podman exec`

This example creates a virtual 127.0.42.0/24 network across a Docker container, a Podman container, and the localhost, using `docker exec` and `podman exec`.

**Step 0: build `bin/norouter` binary** (on Linux)

```console
make
```

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

**Step 4: start the NoRouter "router" process**

```console
./bin/norouter example.yaml
```

**Step 5: connect to `host1` (127.0.42.101, nginx)**

```console
wget -O - http://127.0.42.101:8080
docker exec host1 wget -O - http://127.0.42.101:8080
podman exec host2 wget -O - http://127.0.42.101:8080
```

Make sure nginx's `index.html` ("Welcome to nginx!") is shown.

**Step 6: connect to `host2` (127.0.42.102, Apache httpd)**

```console
wget -O - http://127.0.42.102:8080
docker exec host1 wget -O - http://127.0.42.102:8080
podman exec host2 wget -O - http://127.0.42.102:8080
```

Make sure Apache httpd's `index.html` ("It works!") is shown.

### How it works under the hood

The "router" process of NoRouter launches the following commands and transfer the packets using their stdio streams.

```
/proc/self/exe internal agent \
  --me 127.0.42.100 \
  --other 127.0.42.101:8080 \
  --other 127.0.42.102:8080
```

```
docker exec -i host1 norouter internal agent \
  --me 127.0.42.101 \
  --forward 8080:127.0.0.1:80 \
  --other 127.0.42.102:8080
```

```
podman exec -i host2 norouter internal agent \
  --me 127.0.42.102 \
  --other 127.0.42.101:8080 \
  --forward 8080:127.0.0.1:80
```

`me` is used as a virtual src IP for connecting to `--other <dstIP>:<dstPort>`.

#### stdio protocol

The protocol is still subject to change.
<!-- can we reuse some existing protocol? -->

```
uint32le Len      (includes header fields and Payload but does not include Len itself)
[4]byte  SrcIP
uint16le SrcPort
[4]byte  DstIP
uint16le DstPort
uint16le Proto
uint16le Flags
[]byte   Payload  (without L2/L3/L4 headers at all)
```

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

## TODOs

- Install `norouter` binary to remote hosts automatically?
- Assist generating mTLS certs?
- Add DNS fields to `/etc/resolv.conf` when the file is writable? (writable by default in Docker and Kubernetes)
- Detect port numbers automatically by watching `/proc/net/tcp`, and propagate the information across the cluster automatically?

## Similar projects

- [vdeplug4](https://github.com/rd235/vdeplug4): vdeplug4 can create ad-hoc L2 networks over stdio.
  vdeplug4 is similar to NoRouter in the sense that it uses stdio, but vdeplug4 requires privileges (at least in userNS) for creating TAP devices.
- [telepresence](https://www.telepresence.io/): kube-only and needs privileges
