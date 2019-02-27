# Helm Update Config

This plugin adds `helm update-config` command to Helm CLI. It allows to update config values of an existing release running in the cluster.

## Install

```
helm plugin install https://github.com/bluebosh/helm-update-config
```

## Usage

To change image tag of `smiling-penguin` release:

```
helm update-config smiling-penguin --set-value=image.tag=stable
helm update-config smiling-penguin --reset-value=true
helm update-config smiling-penguin --value=file_path1
```

The plugin will reuse all the values defined in previous releases. If you want to override those you can set `--reset-values` flag the same way you do for `helm upgrade`.

## Maintainers

[@zhanggbj,@edwardstudy,@EmilyEmily](https://github.com/bluebosh)

## Contribute

PRs accepted.

## License

[MIT](LICENSE) © 2019 Emily Jing Liu 
