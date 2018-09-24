## kwt net start

Sets up network access

### Synopsis

Sets up network access

```
kwt net start [flags]
```

### Examples

```
sudo -E kwt net start
```

### Options

```
      --debug                    Set logging level to debug
      --dns-map strings          Domain to IP mapping (can be specified multiple times) (example: 'test.=127.0.0.1')
      --dns-mdns                 Start MDNS server (default true)
  -r, --dns-recursor strings     Recursor (can be specified multiple times)
  -h, --help                     help for start
      --namespace string         Namespace to use to manage networking pod (default "default")
      --remote-ip strings        Additional IP to include for subnet guessing (can be specified multiple times)
      --ssh-host string          SSH server address for forwarding connections (includes port)
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

* [kwt net](kwt_net.md)	 - Network (clean-up, pods, services, start)

