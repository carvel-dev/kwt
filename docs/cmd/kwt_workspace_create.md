## kwt workspace create

Create workspace

### Synopsis

Create workspace

```
kwt workspace create [flags]
```

### Options

```
  -c, --command string              Set command
      --di string                   Set working directory for executing script to particular's input directory
  -d, --directory string            Set working directory for executing script (relative to workspace directory unless is an absolute path)
      --enter                       Enter workspace after create
      --generate-name               Set to generate name (default true)
  -h, --help                        help for create
      --image string                Set image (example: nginx)
      --image-command strings       Set command (can be set multiple times)
      --image-command-arg strings   Set command args (can be set multiple times)
  -i, --input string                Set inputs (format: 'name=local-dir-path' or 'name=local-dir-path:remote-dir-path') (example: knctl=.)
  -n, --namespace string            Specified namespace ($KWT_NAMESPACE or default from kubeconfig)
  -o, --output string               Set outputs (format: 'name=local-dir-path' or 'name=local-dir-path:remote-dir-path') (example: knctl=.)
      --port ints                   Set port (multiple can be specified)
  -p, --privileged                  Set privilged
      --rm                          Remove workspace after execution is finished
      --watch                       Watch and continiously sync inputs
  -w, --workspace string            Specified workspace (default "w")
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

