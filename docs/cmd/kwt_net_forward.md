## kwt net forward

Forward subnet traffic to single port

### Synopsis

Forward subnet traffic to single port

```
kwt net forward [flags]
```

### Options

```
      --debug            Set logging level to debug
  -h, --help             help for forward
  -s, --subnet strings   Subnet (can be specified multiple times)
      --tcp-port int     TCP port destination (default 8080)
      --udp-port int     UDP port destination (default 1234)
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

* [kwt net](kwt_net.md)	 - Network (clean-up, forward, pods, services, start, start-dns)

