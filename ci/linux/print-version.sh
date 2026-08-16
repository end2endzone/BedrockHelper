#!/usr/bin/env bash
set -e

# Get the projet's root directory
PROJECTROOT=$(cd "$(dirname "$0")/../.." && pwd)

cd $PROJECTROOT

# Define the target binary file path based on the environment
TARGET="$PROJECTROOT/bin/bedrock_helper"
if [[ "$CI" == "true" ]]; then
    TARGET="$PROJECTROOT/bin/bedrock_helper-$GOOS-$GOARCH"
    echo "Building on CI/CD server. Changing the target file name to '$TARGET'"
fi

# Show compiled version
$TARGET --version
