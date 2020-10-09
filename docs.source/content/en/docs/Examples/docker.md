---
title: "Docker"
linkTitle: "Docker"
weight: 10
---

Example manifest for Docker:

```yaml
hosts:
  docker:
    cmd: "docker exec -i some-container norouter"
    vip: "127.0.42.101"
    ports: ["8080:127.0.0.1:80"]
```

The `norouter` binary can be installed by using `docker cp`:
```console
$ docker run -d --name foo nginx:alpine
$ docker cp norouter foo:/usr/local/bin
```
