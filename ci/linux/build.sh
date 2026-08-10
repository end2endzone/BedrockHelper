#!/usr/bin/env bash
set -e

# Get the projet's root directory
PROJECTROOT=$(cd "$(dirname "$0")/../.." && pwd)

cd $PROJECTROOT/src

VERSION=$(cat $PROJECTROOT/VERSION)
COMMIT=$(git rev-parse --short HEAD)
DATE=$(date +%F)

# GOOS
if [[ -n "${GOOS}" ]]; then
    echo "GOOS is set to: $GOOS"
else
    # Detect OS type
    #OS_TYPE=$(uname -s | tr '[:upper:]' '[:lower:]')
    ## Truncate OS_TYPE when running on specific OS
    #case "$OS_TYPE" in
    #linux)      GOOS="linux" ;;
    #darwin)     GOOS="darwin" ;;
    #*mingw64*|
    #*mingw32*|
    #*msys*|
    #*cygwin*)   GOOS="windows" ;;
    #*)          GOOS="unknown" ;;
    #esac

    GOOS=$(go env GOHOSTOS)
    echo "GOOS is not set and is forced to: $GOOS"
fi

# GOARCH
if [[ -n "${GOARCH}" ]]; then
    echo "GOARCH is set to: $GOARCH"
else
    GOARCH=$(go env GOHOSTARCH)
    echo "GOARCH is not set and is forced to: $GOARCH"
fi

# Ensures your binary does not depend on host operating system C libraries, making the binary completely portable.
CGO_ENABLED=0

# Point this to the exact package path where your main function lives
PKG="main"

echo "Building bedrock_helper-$GOOS-$GOARCH version $VERSION..."

# Run build from the root, pointing to the main package directory
go build -ldflags "-X '${PKG}.Version=${VERSION}' -X '${PKG}.CommitHash=${COMMIT}' -X '${PKG}.BuildDate=${DATE}'" -o $PROJECTROOT/bin/bedrock_helper-$GOOS-$GOARCH ./cmd/bedrock_helper

echo "Build complete!"
