## kwt resource get

Get resource

### Synopsis

Get resource

```
kwt resource get [flags]
```

### Options

```
  -h, --help                   help for get
  -n, --namespace string       Specified namespace ($KWT_NAMESPACE or default from kubeconfig)
  -r, --resource string        Specified resource name (format: 'name')
  -t, --resource-type string   Specified resource type (format: 'group/version/resource')
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

* [kwt resource](kwt_resource.md)	 - Resource (delete, get, list)

