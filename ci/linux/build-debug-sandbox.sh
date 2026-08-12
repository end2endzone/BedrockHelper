#!/usr/bin/env bash
set -e

# Get the projet's root directory
PROJECTROOT=$(cd "$(dirname "$0")/../.." && pwd)

cd $PROJECTROOT

rm -rf .debug-sandbox
mkdir -p .debug-sandbox
cp -r src/testdata .debug-sandbox
cp src/testdata/addons/foobar.mcaddon .debug-sandbox/testdata/servers/server_with_multiple_packs/foobar.mcaddon

# Testing installing multiple addon at once
cp -r src/testdata/servers/server_empty .debug-sandbox/testdata/servers/server_with_uninstalled_addons
cp -r src/testdata/addons               .debug-sandbox/testdata/servers/server_with_uninstalled_addons/my-addons-collection

echo "Sandbox ready!"
