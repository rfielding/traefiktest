Traefik Test
===========

The `traefik` reverse proxy is typically used in Docker or Rancher setups.
But it is possible to run it in a completely standalone manner. The `traefik` binary

- has a static config, in `traefik.yaml`, which is the config given on startup
- has a dyndmic config, in `dynamic.yaml`, which is more config that listens for
  a hot reload of the data. It is meant that `dynamic.yaml` is manipulated by tools
  that can calculate changes to `dynamic.yaml`.  For example, when the docker
  socket is listeneed to, the dynamic yaml is rewritten to keep up.  This is
  how cluster members get added and removed.

Simple Test
==========

> Assuming that `traefik` binary is in your path, and a Go compiler is installed.

We want a pair of load-balanced services at http://localhost:8000

- `/app1` proxies back to 
  - http://localhost:1111 
  - http://localhost:3333
- `/app2` proxies back to
  - http://localhost:2222
  - http://localhost:4444

Note that the load balancer does not automatically remove the down service fropm the list.
This means that something needs to do a health check and _remove_ a down host from the list
that it round robin between.

> ie: if 3333 goes down, then /app1 alternates between a 200 for service on 1111, and a 503 on 3333 
 
Run 4 different servers under traefik to test out load balance and reverse proxy

```
./build
```

Then browser urls

- The Admin UI: http://localhost:8080
- Server 1 http://localhost:8000/app1
- Server 2 http://localhost:8000/app2

The services on port 1111,3333 and 2222,4444 are going up and down constantly. When all are down, we have an outage. The reconcile.go program keeps rewriting the cluster members to keep up with what actually exists.
