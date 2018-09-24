## kwt net start-dns

Start DNS server and redirects system DNS resolution to it

### Synopsis

Start DNS server and redirects system DNS resolution to it

```
kwt net start-dns [flags]
```

### Examples

```
sudo -E kwt net start-dns
```

### Options

```
      --debug              Set logging level to debug
  -h, --help               help for start-dns
      --map strings        Domain to IP mapping (can be specified multiple times) (example: 'test.=127.0.0.1')
      --mdns               Start MDNS server (default true)
  -r, --recursor strings   Recursor (can be specified multiple times)
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

