---
title: "Kubernetes"
linkTitle: "Kubernetes"
weight: 14
---

Example manifest for Kubernetes:

```yaml
hosts:
  kube:
    cmd: "kubectl --context=some-context exec -i some-pod -- norouter"
    vip: "127.0.42.102"
    ports: ["8080:127.0.0.1:80"]
# Writing /etc/hosts is possible on most Docker and Kubernetes containers
    writeEtcHosts: true
```

The `norouter` binary can be installed by using `kubectl cp`:
```console
$ kubectl run --image=nginx:alpine --restart=Never nginx
$ kubectl cp norouter nginx:/usr/local/bin
```

## Multi-cluster

To connect multiple Kubernetes clusters, pass `--context` arguments to `kubectl`.

e.g. To connect GKE, AKS, and your laptop:

```yaml
hosts:
  laptop:
    vip: "127.0.42.100"
  nginx-on-gke:
    cmd: "kubectl --context=gke_myproject-12345_asia-northeast1-c_my-gke exec -i nginx -- norouter"
    vip: "127.0.42.101"
    ports: ["8080:127.0.0.1:80"]
  httpd-on-aks:
    cmd: "kubectl --context=my-aks exec -i httpd -- norouter"
    vip: "127.0.42.102"
    ports: ["8080:127.0.0.1:80"]
```
