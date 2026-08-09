#!/usr/bin/env bash
set -e

# Get the projet's root directory
PROJECTROOT=$(cd "$(dirname "$0")/../.." && pwd)

cd $PROJECTROOT/src

VERSION=$(cat $PROJECTROOT/VERSION)
COMMIT=$(git rev-parse --short HEAD)
DATE=$(date +%F)

# Point this to the exact package path where your main function lives
PKG="main"

echo "Building BedrockHelper version $VERSION..."

# Run build from the root, pointing to the main package directory
go build -ldflags "-X '${PKG}.Version=${VERSION}' -X '${PKG}.CommitHash=${COMMIT}' -X '${PKG}.BuildDate=${DATE}'" -o $PROJECTROOT/bin/bedrock_helper ./cmd/bedrock_helper

echo "Build complete!"
