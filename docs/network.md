## Network Commands

`kwt net start` commands provide a way to access Kubernetes services and pods via DNS or IPs directly from your workstation. Only OS X and Linux is currently supported. 

Possible uses:

- access HTTP and HTTPs applications
- access TCP services such as Redis
- access applications via their Kubernetes DNS address or overlay IP address
- remap some domains to different IPs to aid testing (useful when )

To get started, make sure your Kubernetes cluster is ready

```bash
my-laptop $ kubectl get node
```

Let's also install Kubernetes [example guestbook application](https://github.com/kubernetes/examples/tree/master/guestbook)

```bash
my-laptop $ wget -O- https://raw.githubusercontent.com/kubernetes/examples/3a2f276a561d877e0f93986829568eeea98d2327/guestbook/all-in-one/guestbook-all-in-one.yaml | kubectl apply -f -
```

Once complete run `kwt net svc` to list networking details about services in current namespace

```bash
my-laptop $ kwt net svc

Services in namespace 'default'

Name          Internal DNS                            Cluster IP     Ports
frontend      frontend.default.svc.cluster.local      10.19.247.124  80/tcp
redis-master  redis-master.default.svc.cluster.local  10.19.244.101  6379/tcp
redis-slave   redis-slave.default.svc.cluster.local   10.19.245.130  6379/tcp

4 services

Succeeded
```

Next, run below command in a separate terminal

```bash
my-laptop $ sudo -E kwt net start
# ...snip...
02:29:52PM: info: ForwardingProxy: Ready!
```

It may take ~1 min to start for the first time as it creates a pod in your Kubernetes cluster. (`-E` in `sudo -E` is for propagating environment variables such as `KUBECONFIG`.)

You now should be able to use regular tools such as `curl` or your web browser to talk to services running in your Kubernetes env.

```bash
my-laptop $ curl -vvv frontend.default.svc.cluster.local

* Rebuilt URL to: frontend.default.svc.cluster.local/
* Hostname was NOT found in DNS cache
*   Trying 10.19.247.124...
* Connected to frontend.default.svc.cluster.local (10.19.247.124) port 80 (#0)
> GET / HTTP/1.1
> User-Agent: curl/7.37.1
> Host: frontend.default.svc.cluster.local
> Accept: */*
>
< HTTP/1.1 200 OK
< Date: Fri, 28 Sep 2018 23:18:56 GMT
* Server Apache/2.4.10 (Debian) PHP/5.6.20 is not blacklisted
< Server: Apache/2.4.10 (Debian) PHP/5.6.20
< Last-Modified: Wed, 09 Sep 2015 18:35:04 GMT
< ETag: "399-51f54bdb4a600"
< Accept-Ranges: bytes
< Content-Length: 921
< Vary: Accept-Encoding
< Content-Type: text/html
<
<html ng-app="redis">
  <head>
    <title>Guestbook</title>
# ...snip...
```

Let's also connect to Redis service that's used for storing guestbook data

```bash
my-laptop $ redis-cli -h redis-master.default.svc.cluster.local

redis-master.default.svc.cluster.local:6379> get messages
",hello"
```

### Cheatsheet

Start networking access and guess as much configuration as possible

```bash
sudo -E kwt net start
```

Start networking access, allowing access to two subnets

```bash
sudo -E kwt net start --subnet 10.19.247.0/24 --subnet 10.19.248.0/24
```

Start networking access, and configure `example.com` or anything under it (such as `test.t.example.com`) to resolve to `127.0.0.1` (Hint: useful with `knctl` to forward requests to Knative ingress without official DNS changes)

```bash
sudo -E kwt net start --dns-map example.com=127.0.0.1
```

Show services in the current/specified namespace

```bash
kwt net svc [-n ns1]
```

Show pods in the current/specified namespace

```bash
kwt net pods [-n ns1]
```

Clean up on-cluster resources taken by `kwt net` (one pod and two secrets)

```bash
kwt net clean-up
```

### Other tools

- [`kubectl proxy` command](https://kubernetes.io/docs/tasks/access-application-cluster/access-cluster/) acts as a reverse proxy. It rewrites HTTP URLs and does not offer access to TCP services making it not viable for some use cases.

- [`kubectl port-forward` command](https://kubernetes.io/docs/tasks/access-application-cluster/port-forward-access-application-cluster/) provides a simple way to forward traffic from a local port to some port inside Kubernetes cluster.

- [`sshuttle`](https://github.com/sshuttle/sshuttle) is a generic tool that is used to forward traffic transparently from your local machine to remote machine via SSH. Unfortunately sshuttle is written in Python making it somewhat challenging to package it as a single easy to install binary. `kwt net` firewall configuration (iptables, pfctl) is almost exactly same as sshuttle's; however, kwt currently relies on sshd's SOCKS5 support for making remote connections.

- [`telepresence`](https://github.com/telepresenceio/telepresence/) is a tool that wraps sshuttle and makes it more Kubernetes native.
