## kwt workspace install

Install predefined application into workspace

### Synopsis

Install predefined application into workspace

```
kwt workspace install [flags]
```

### Options

```
      --chrome             Install Google Chrome
      --desktop            Configure X11 and VNC access
      --docker             Install Docker
      --firefox            Install Firefox
      --go1x               Install Go 1.x
  -h, --help               help for install
  -n, --namespace string   Specified namespace ($KWT_NAMESPACE or default from kubeconfig)
      --sublime            Install Sublime Text
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

