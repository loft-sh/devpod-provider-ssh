# DevPod SSH Provider

[![Join us on Slack!](docs/static/media/slack.svg)](https://slack.loft.sh/) [![Open in DevPod!](https://devpod.sh/assets/open-in-devpod.svg)](https://devpod.sh/open#https://github.com/loft-sh/devpod-provider-ssh)

This repository hosts the default SSH provider configuration used in DevPod.

## Usage

To add this SSH provider from the CLI, use the `provider add` command. For example:

```shell
devpod provider add ssh
```

## Compatibility

We only support Linux machine as remote hosts.

### Windows

There are known issues with the default windows SSH installation in some setups. If you're unable to connect to your host by default,
try to enable the `USE_BUILTIN_SSH` option
```shell
devpod provider add ssh --option USE_BUILTIN_SSH=true
# or if already installed
devpod provider set-options ssh --option USE_BUILTIN_SSH=true
```

This forces the provider to use the builtin SSH client over the one accessible in your shell. 
You will need to add the identities file manually to your SSH config in case it's not the default key:
```ssh
Host my-domain.com
    User my-user 
    IdentityFile ~/.my-dir/my-key
```

## Options

This provider has the following options:

| NAME            | REQUIRED | DESCRIPTION                                                | DEFAULT           |
|-----------------|----------|------------------------------------------------------------|-------------------|
| HOST            | true     | The SSH Host to connect to. Example: my-user@my-domain.com |                   |
| AGENT_PATH      | false    | The path where to inject the DevPod agent to.              | /tmp/devpod/agent |
| DOCKER_PATH     | false    | The path of the docker binary.                             | docker            |
| EXTRA_FLAGS     | false    | Extra flags to pass to the SSH command.                    |                   |
| PORT            | false    | The SSH port to use.                                       | 22                |
| USE_BUILTIN_SSH | false    | Use the builtin SSH package.                               | false             |

# Extra

For more detail, see the [DevPod Documentation](https://devpod.sh/docs/managing-providers/what-are-providers).
