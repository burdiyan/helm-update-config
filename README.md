# Helm Update Config

This plugin adds `helm update-config` command to Helm CLI. It allows to update config values of an existing release running in the cluster.

## Install

```
helm plugin install https://github.com/bluebosh/helm-update-config
```

## Usage

To change image tag of `smiling-penguin` release:

```
helm update-config "cf.20190214.142635" --set-value image.tag=stable,kube.diego_cell_size=1
helm update-config "cf.20190214.142635" --values file_path1
helm update-config "cf.20190214.142635" --set-value image.tag=stable --values=file_path1
```

For the last sample command, the plugin will merge the key/value pairs specified in both --set-value and --value file, if there are different keys specified in both --set-value and --value file, then --set-value will override the value in --value file.


## Maintainers

[@zhanggbj,@edwardstudy,@EmilyEmily](https://github.com/bluebosh)

## Contribute

PRs accepted.
If you want to make a new release, either set GitHub API Token via GITHUB_TOKEN env:
```
$ export GITHUB_TOKEN="....."
```
and then run ghr command with "-t ${GITHUB_TOKEN}" parameter, or set github.token in ./gitconfig:
```
$ cat ~/.gitconfig
......
[github]
	token = ******
```

## License

[MIT](LICENSE) © 2019 Emily Jing Liu 
