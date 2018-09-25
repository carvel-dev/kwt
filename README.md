# kwt

Kubernetes Workstation Tools CLI (`kwt`) provides helpful set of commands for application developers to develop applications with Kubernetes.

Grab pre-built binaries from the [Releases page](https://github.com/cppforlife/kwt/releases).

## Network Commands

You may want to access Kubernetes services and pods via DNS or IPs directly from your workstation (OS X for example). To do so run below command in a separate terminal. You now should be able to use your regular tools such as `curl` or your web browser to talk to services running in your Kubernetes env.

```bash
sudo -E kwt net start
# ...snip...
02:29:52PM: info: ForwardingProxy: Ready!
```

(Only OS X and Linux currently supported. It may take ~1 min to start for the first time as it creates a pod in your Kubernetes cluster.)

If you want a handy way to find available networking components and their addresses, try:

```bash
kwt net svc [-n ns1]
kwt net pods [-n ns1]
```

## Workspace Commands

A lot of times you may want a "scratch" container to run some scripts or experiment with different commands but without affecting your local machine or consuming its resources. If you have a Kubernetes cluster, you can use following commands to do so quickly:

```bash
kwt workspace create -i mydir=. -c './mydir/script.sh' --rm
# or kwt w c ... for short
```

Above command used `--input` (`-i`) to upload local current directory.

Alterntaively if you want empty container with a shell:

```bash
kwt workspace create
# ...snip... obtain workspace name
kwt workspace enter -w <some-name>
# or
kwt workspace create --enter
```

In some cases you may want to run additional scripts in an existing workspaces:

```bash
kwt workspace ls
kwt workspace run -w <some-name> -i kwt=.:gopath/src/github.com/cppforlife/kwt -c './gopath/src/github.com/cppforlife/kwt/ci/unit-tests.sh'
```

Once done, delete your workspace

```bash
kwt workspace delete -w <some-name>
```
