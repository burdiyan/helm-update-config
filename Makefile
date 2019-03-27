build:
	@gox -osarch "darwin/amd64 linux/amd64" -output "dist/{{.Dir}}_{{.OS}}_{{.Arch}}"

release:
	@ghr -u bluebosh -r helm-update-config -replace `git describe --tags` dist/

.PHONY: build release