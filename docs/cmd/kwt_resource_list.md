## kwt resource list

List all resources

### Synopsis

List all resources

```
kwt resource list [flags]
```

### Options

```
      --add-includes string     Filter by content in resource annotation keys
  -h, --help                    help for list
      --label-includes string   Filter by content in resource label keys
      --name-includes string    Filter by content in resource name
  -n, --namespace string        Specified namespace ($KWT_NAMESPACE or default from kubeconfig)
  -r, --resource string         Specified resource name (format: 'name')
  -t, --resource-type string    Specified resource type (format: 'group/version/resource')
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

