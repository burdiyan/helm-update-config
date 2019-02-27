# Helm Update Config

This plugin adds `helm update-config` command to Helm CLI. It allows to update config values of an existing release running in the cluster.

## Install

```
helm plugin install https://github.com/bluebosh/helm-update-config
```

## Usage

To change image tag of `smiling-penguin` release:

```
helm update-config "cf.20190214.142635" --set-value=image.tag=stable,kube.diego_cell_size=1
helm update-config "cf.20190214.142635" --value=file_path1
helm update-config "cf.20190214.142635" --set-value=image.tag=stable --value=file_path1
```

For the last sample command, the plugin will merge the key/value pairs in set-value and --value

The plugin will reuse all the values defined in previous releases. If you want to override those you can set `--reset-values` flag the same way you do for `helm upgrade`.

## Maintainers

[@zhanggbj,@edwardstudy,@EmilyEmily](https://github.com/bluebosh)

## Contribute

PRs accepted.

## License

[MIT](LICENSE) © 2019 Emily Jing Liu 
