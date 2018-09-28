## Workspace Commands (work in progress!)

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

### Input flag format

```bash
# Upload current directory to /tmp/build/x/kwt
kwt w s -i kwt=.

# Upload current directory to /tmp/build/x/kwt/gopath/src/github.com/cppforlife/kwt
kwt w s -i kwt=.:gopath/src/github.com/cppforlife/kwt

# Upload current directory to remote abs path /tmp/kwt
kwt w s -i kwt=.:/tmp/kwt
```

### Output flag format

```bash
# Download /tmp/build/x/kwt to current directory
kwt w s -o kwt=.

# Download gopath/src/github.com/cppforlife/kwt to current directory
kwt w s -o kwt=.:gopath/src/github.com/cppforlife/kwt

# Download /tmp/build/x/kwt to /tmp/kwt
kwt w s -o kwt=/tmp/kwt
```
