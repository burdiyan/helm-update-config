#!/usr/bin/env bash

set -euf -o pipefail

PLUGIN_VERSION=${PLUGIN_VERSION:-"0.4.1"}

file="${HELM_PLUGIN_DIR:-"$(helm home)/plugins/helm-update-config"}/bin/helm-update-config"

mkdir -p $(dirname ${file})

os=$(uname -s | tr '[:upper:]' '[:lower:]')
url="https://github.com/bluebosh/helm-update-config/releases/download/v${PLUGIN_VERSION}/helm-update-config_${os}_amd64"

if command -v wget; then
  wget -O "${file}"  "${url}"
elif command -v curl; then
  curl -o "${file}" -L "${url}"
fi

chmod +x "${file}"
