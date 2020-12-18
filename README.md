# kwt

- Slack: [#carvel in Kubernetes slack](https://slack.kubernetes.io)
- Install: grab pre-built binaries from the [Releases page](https://github.com/k14s/kwt/releases) or [Homebrew Carvel tap](https://github.com/vmware-tanzu/homebrew-carvel)
- Status: Experimental

`kwt` (Kubernetes Workstation Tools) (pronounced spelled out) CLI provides helpful set of commands for application developers to develop applications with Kubernetes.

- See [Network](docs/network.md) to find out how to make Kubernetes services and pods accessible to your local machine
- See [Workspace](docs/workspace.md) to find out how to conveniently manage temporary containers that could be used for debugging, experimentation etc.

## Development

```bash
./hack/build.sh
export KWT_E2E_NAMESPACE=kwt-test
sudo -E ./hack/test-all.sh
```
