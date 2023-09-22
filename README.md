![logo](docs/CarvelLogo.png)

# kwt

- Slack: [#carvel in Kubernetes slack](https://slack.kubernetes.io)
- Install: grab pre-built binaries from the [Releases page](https://github.com/carvel-dev/kwt/releases) or [Homebrew Carvel tap](https://github.com/carve-dev/carvel)
- Status: Experimental

`kwt` (Kubernetes Workstation Tools) (pronounced spelled out) CLI provides helpful set of commands for application developers to develop applications with Kubernetes.

- See [Network](docs/network.md) to find out how to make Kubernetes services and pods accessible to your local machine
- See [Workspace](docs/workspace.md) to find out how to conveniently manage temporary containers that could be used for debugging, experimentation etc.

### Join the Community and Make Carvel Better
Carvel is better because of our contributors and maintainers. It is because of you that we can bring great software to the community.
Please join us during our online community meetings. Details can be found on our [Carvel website](https://carvel.dev/community/).

You can chat with us on Kubernetes Slack in the #carvel channel and follow us on Twitter at @carvel_dev.

Check out which organizations are using and contributing to Carvel: [Adopter's list](https://github.com/carvel-dev/carvel/blob/master/ADOPTERS.md)

## Development

```bash
./hack/build.sh
export KWT_E2E_NAMESPACE=kwt-test
sudo -E ./hack/test-all.sh
```
