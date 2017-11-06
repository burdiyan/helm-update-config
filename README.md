# Helm Update Config

This plugin adds `helm update-config` command to Helm CLI. It allows to update config values of an existing release running in the cluster.

## Install

```
helm plugin install https://github.com/burdiyan/helm-update-config
```

## Usage

To change image tag of `smiling-penguin` release:

```
helm update-config smiling-penguin --set=image.tag=stable
```

The plugin will reuse all the values defined in previous releases. If you want to override those you can set `--reset-values` flag the same way you do for `helm upgrade`.

## Maintainers

[@burdiyan](https://github.com/burdiyan)

## Contribute

PRs accepted.

## License

[MIT](LICENSE) © 2017 Alexandr Burdiyan
