## kwt workspace sync

Sync workspace

### Synopsis

Sync workspace

```
kwt workspace sync [flags]
```

### Options

```
  -h, --help               help for sync
  -i, --input string       Set inputs (format: 'name=local-dir-path' or 'name=local-dir-path:remote-dir-path') (example: knctl=.)
  -n, --namespace string   Specified namespace ($KWT_NAMESPACE or default from kubeconfig)
  -o, --output string      Set outputs (format: 'name=local-dir-path' or 'name=local-dir-path:remote-dir-path') (example: knctl=.)
      --watch              Watch and continiously sync inputs
  -w, --workspace string   Specified workspace
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

* [kwt workspace](kwt_workspace.md)	 - Workspace (add-alt-name, create, delete, enter, install, list, run, sync)

