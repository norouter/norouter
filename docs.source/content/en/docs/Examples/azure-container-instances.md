---
title: "Azure Container Instances"
linkTitle: "Azure Container Instances"
weight: 80
---
{{% alert %}}
**Note**:

This article is about Azure Container Instances (ACI), not about Azure Kubernetes Service (AKS).

To see how to use NoRouter with AKS, see [Kubernetes example](../kubernetes/).
{{% /alert %}}

In [Kubernetes example](../kubernetes/), we used `kubectl exec` as the stdio connection.

However, unlike `kubectl exec`, NoRouter cannot support `az container exec` currently because:
- No support for stdin without tty: https://github.com/Azure/azure-cli/issues/15225
- No support for appending command arguments: https://docs.microsoft.com/en-us/azure/container-instances/container-instances-exec#restrictions
- Extra TTY escape sequence on busybox: https://github.com/Azure/azure-cli/issues/6537

A workaround is to inject an SSH sidecar into an Azure container group, and use `ssh` instead of `az container exec`.

## SSH Sidecar

To create an Azure container group with an SSH sidecar, create a `Microsoft.ContainerInstance/containerGroups` object like this:

```yaml
type: Microsoft.ContainerInstance/containerGroups
api-version: 2019-12-01
name: nginx-with-ssh
properties:
  osType: Linux
  ipAddress:
    type: Public
# Set an arbitrary unique name
    dnsNameLabel: "********"
    ports:
    - port: 2222
      protocol: TCP
  containers:
  - name: nginx
    properties:
      image: nginx
      resources:
        requests:
          cpu: 1.0
          memoryInGB: 2.0
  - name: ssh
    properties:
# Dockerfile: https://github.com/linuxserver/docker-openssh-server
      image: linuxserver/openssh-server
      environmentVariables:
# Set an arbitrary name
      - name: USER_NAME
        value: "johndoe"
# Set the content of ~/.ssh/id_rsa.pub
      - name: PUBLIC_KEY
        value: "ssh-rsa AAAB******** ********@********"
      ports:
      - port: 2222
        protocol: TCP
      resources:
        requests:
          cpu: 0.5
          memoryInGB: 0.5
```

At least, you need to set the following fields to your own values:
- `.properties.ipAddress.dnsNameLabel`: An arbitrary unique string. When the value is like "example" and the region is "westus", the FQDN will be like "example.westus.azurecontainer.io".
- `.properties.containers[?(@.name=="ssh")].environmentVariables[?(@.name=="PUBLIC_KEY")].value`: The content of `~/.ssh/id_rsa.pub`

See also ["YAML reference: Azure Container Instances" (`docs.microsoft.com`)](https://docs.microsoft.com/en-us/azure/container-instances/container-instances-reference-yaml).

Save the object as `nginx-with-ssh.yaml`.
Then create the container group with `az container create`, and inspect the FQDN with `az container show` as follows:
```console
$ az container create -f nginx-with-ssh.yaml
$ az container show -n nginx-with-ssh -o json | jq -r .ipAddress.fqdn
<unique-name>.<region>.azurecontainer.io
```

After the container group booted up, install NoRouter to the SSH sidecar:

```console
$ norouter show-installer | ssh -p 2222 johndoe@<unique-name>.<region>.azurecontainer.io
...
Successfully installed /config/bin/norouter (version 0.4.0)
```

NoRouter manifest can be written like this:
```yaml
hosts:
  local:
    vip: "127.0.42.100"
  aci:
    cmd: "ssh -p 2222 johndoe@<unique-name>.<region>.azurecontainer.io -- /config/bin/norouter"
    vip: "127.0.42.101"
    ports: ["8080:127.0.0.1:80"]
```

Make sure http://127.0.42.101:8080 is forwarded to the `nginx` container inside the `nginx-with-ssh` container group.
```console
$ norouter norouter.yaml &
$ curl http://127.0.42.101:8080
...
<title>Welcome to nginx</title>
...
```
