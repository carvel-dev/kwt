## kwt registry install

Install Docker registry into your cluster

### Synopsis

Install Docker registry into your cluster

```
kwt registry install [flags]
```

### Examples

```
kwt registry install
```

### Options

```
      --debug              Set logging level to debug
  -h, --help               help for install
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

