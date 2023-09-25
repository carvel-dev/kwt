## kwt net start

Sets up network access

### Synopsis

Sets up network access

```
kwt net start [flags]
```

### Examples

```

  # Detect settings automatically
  sudo -E kwt net start

  # Provide predefined set of subnets to proxy
  sudo -E kwt net start --subnet 10.19.247.0/24 --subnet 10.19.248.0/24

  # Redirect all example.com and its subdomains to localhost
  sudo -E kwt net start --dns-map example.com=127.0.0.1

  # Dynamically configure DNS mappings
  sudo -E kwt net start --dns-map-exec='knctl dns-map'

```

### Options

```
      --debug                    Set logging level to debug
      --dns-map strings          Domain to IP or Kubernetes DNS mapping (can be specified multiple times) (example: 'test.=127.0.0.1', 'custom.=kubernetes')
      --dns-map-exec strings     Domain to IP mapping command to execute periodically (can be specified multiple times) (example: 'knctl dns-map')
      --dns-mdns                 Start MDNS server (default true)
  -r, --dns-recursor strings     Recursor (can be specified multiple times)
  -h, --help                     help for start
  -n, --namespace string         Namespace to use to manage networking pod (default "default")
      --remote-ip strings        Additional IP to include for subnet guessing (can be specified multiple times)
      --ssh-host string          SSH server address for forwarding connections (includes port)
      --ssh-image string         Image URL to use for starting OpenSSH on K8s (default "ghcr.io/carvel-dev/kwt/sshd@sha256:b47888724e3d891a3c8cb15155f9a434468b316c0e00a96e920fb5d1121cc4b0")
      --ssh-private-key string   Private key for connecting to SSH server (PEM format)
      --ssh-user string          SSH server username
  -s, --subnet strings           Subnet, if specified subnets will not be guessed automatically (can be specified multiple times)
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

