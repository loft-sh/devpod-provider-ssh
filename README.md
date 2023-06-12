# DevPod SSH Provider

This repository hosts the default SSH provider configuration used in DevPod.

## Usage

To add this SSH provider from the CLI, use the `provider add` command. For example:

```shell
# be sure to set $CURRENT_VERSION to an appropriate release tag from this repo
devpod provider add https://github.com/loft-sh/devpod-provider-ssh/releases/download/$CURRENT_VERSION/provider.yaml
```

For more detail, see the [DevPod Documentation](https://devpod.sh/docs/managing-providers/what-are-providers).