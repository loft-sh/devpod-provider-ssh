# DevPod SSH Provider

[![Join us on Slack!](docs/static/media/slack.svg)](https://slack.loft.sh/) [![Open in DevPod!](https://devpod.sh/assets/open-in-devpod.svg)](https://devpod.sh/open#https://github.com/loft-sh/devpod-provider-ssh)

This repository hosts the default SSH provider configuration used in DevPod.

## Usage

To add this SSH provider from the CLI, use the `provider add` command. For example:

```shell
# be sure to set $CURRENT_VERSION to an appropriate release tag from this repo
devpod provider add https://github.com/loft-sh/devpod-provider-ssh/releases/download/$CURRENT_VERSION/provider.yaml
```

## Compatibility

We only support Linux machine as remote hosts.

# Extra

For more detail, see the [DevPod Documentation](https://devpod.sh/docs/managing-providers/what-are-providers).
