# kwt

- Slack: [#k14s in Kubernetes slack](https://slack.kubernetes.io)
- Install: grab pre-built binaries from the [Releases page](https://github.com/k14s/kwt/releases)

`kwt` (Kubernetes Workstation Tools) (pronounced spelled out) CLI provides helpful set of commands for application developers to develop applications with Kubernetes.

- See [Network](docs/network.md) to find out how to make Kubernetes services and pods accessible to your local machine
- See [Workspace](docs/workspace.md) to find out how to conveniently manage temporary containers that could be used for debugging, experimentation etc.

## Development

```bash
./hack/build.sh
sudo -E ./hack/test-all.sh
```
