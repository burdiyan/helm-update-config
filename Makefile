build:
	@gox -osarch "darwin/amd64 linux/amd64" -output "dist/{{.Dir}}_{{.OS}}_{{.Arch}}" helm-update-config

release:
	@ghr -t ${GITHUB_TOKEN} -u EmilyEmily -r helm-update-config --replace `git describe --tags` dist/

.PHONY: build release