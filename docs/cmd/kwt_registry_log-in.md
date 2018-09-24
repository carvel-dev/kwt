## kwt registry log-in

Log in Docker daemon to use cluster registry

### Synopsis

Log in Docker daemon to use cluster registry

```
kwt registry log-in [flags]
```

### Examples

```
kwt registry l
```

### Options

```
  -h, --help               help for log-in
      --namespace string   Namespace to use to find/install cluster registry (default "kwt-cluster-registry")
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

* [kwt registry](kwt_registry.md)	 - Registry (info, install, log-in, uninstall)

