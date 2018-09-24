## kwt workspace run

Run executable within a workspace

### Synopsis

Run executable within a workspace

```
kwt workspace run [flags]
```

### Options

```
  -c, --command string     Set command
      --di string          Set working directory for executing script to particular's input directory
  -d, --directory string   Set working directory for executing script (relative to workspace directory unless is an absolute path)
      --enter              Enter workspace after create
  -h, --help               help for run
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

* [kwt workspace](kwt_workspace.md)	 - Workspace (create, delete, enter, list, run, sync)

