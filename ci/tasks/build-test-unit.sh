#!/bin/sh

set -e

workspace="${GOPATH}/src/github.com/bluebosh/helm-update-config"
mkdir -p ${workspace}
cp -r helm-update-config/* ${workspace}
cd ${workspace}
    
./bin/tools
./bin/unit_test

echo "Unit test finished."
