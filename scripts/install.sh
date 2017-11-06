#!/usr/bin/env bash

set -euf -o pipefail

PLUGIN_VERSION=${PLUGIN_VERSION:-"0.1.0"}

file="${HELM_PLUGIN_DIR:-"$(helm home)/plugins/helm-update-config"}/bin/helm-update-config"
os=$(uname -s | tr '[:upper:]' '[:lower:]')
url="https://github.com/burdiyan/helm-update-config/releases/download/v${PLUGIN_VERSION}/helm-edit_${os}_amd64"

if command -v wget; then
  wget -O "${file}"  "${url}"
elif command -v curl; then
  curl -o "${file}" "${url}"
fi

chmod +x "${file}"
