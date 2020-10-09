---
title: "Azure Container Instances"
linkTitle: "Azure Container Instances"
weight: 80
---


`az container exec` can't be supported currently because:
- No support for stdin without tty: https://github.com/Azure/azure-cli/issues/15225
- No support for appending command arguments: https://docs.microsoft.com/en-us/azure/container-instances/container-instances-exec#restrictions
- Extra TTY escape sequence on busybox: https://github.com/Azure/azure-cli/issues/6537

A workaround is to inject an SSH sidecar into an Azure container group, and use `ssh` instead of `az container exec`.

(To be documented)
