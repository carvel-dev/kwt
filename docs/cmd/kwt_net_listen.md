## kwt net listen

Redirect incoming service traffic to a local port

### Synopsis

Redirect incoming service traffic to a local port

```
kwt net listen [flags]
```

### Examples

```

  # Create service 'svc1' and forward its port 80 to localhost:80
  kwt net listen --service svc1

  # Create service 'svc1' and forward its port 8080 to localhost:8081
  kwt net listen -r 8080 -l localhost:8081 --service svc1

  # TODO Create service and forward it to local listening process
  kwt net l --service svc1 --local-process foo

```

### Options

```
      --debug                    Set logging level to debug
  -h, --help                     help for listen
  -l, --local string             Local address (example: 80, localhost:80) (default "localhost:80")
  -n, --namespace string         Specified namespace ($KWT_NAMESPACE or default from kubeconfig)
  -r, --remote string            Remote address (example: 80) (default "80")
  -s, --service string           Service to create or update for incoming traffic
      --service-type string      Service type to set if creating service (default "ClusterIP")
      --ssh-host string          SSH server address for forwarding connections (includes port)
      --ssh-image string         Image URL to use for starting OpenSSH on K8s (default "ghcr.io/carvel-dev/kwt/sshd@sha256:b47888724e3d891a3c8cb15155f9a434468b316c0e00a96e920fb5d1121cc4b0")
      --ssh-private-key string   Private key for connecting to SSH server (PEM format)
      --ssh-user string          SSH server username
```

### Options inherited from parent commands

```
      --column strings              Filter to show only given columns
      --json                        Output as JSON
      --kubeconfig string           Path to the kubeconfig file ($KWT_KUBECONFIG or $KUBECONFIG)
      --kubeconfig-context string   Kubeconfig context override ($KWT_KUBECONFIG_CONTEXT)
      --no-color                    Disable colorized output
      --non-interactive             Don't ask for user input
      --tty                         Force TTY-like output
```

### SEE ALSO

* [kwt net](kwt_net.md)	 - Network (clean-up, listen, pods, services, start)

