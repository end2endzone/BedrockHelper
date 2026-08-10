#!/usr/bin/env bash
set -e

# Get the projet's root directory
PROJECTROOT=$(cd "$(dirname "$0")/../.." && pwd)

cd $PROJECTROOT/src

VERSION=$(cat $PROJECTROOT/VERSION)
COMMIT=$(git rev-parse --short HEAD)
DATE=$(date +%F)
CPU_ARCHITECTURE=$(uname -m)
OS_TYPE=$(uname -s | tr '[:upper:]' '[:lower:]')

# Truncate OS_TYPE when running on specific OS
case "$OS_TYPE" in
  linux)        OS_TYPE="linux" ;;
  darwin)       OS_TYPE="macos" ;;
  *mingw64*)    OS_TYPE="mingw64" ;;
  *mingw32*)    OS_TYPE="mingw32" ;;
  *msys*)       OS_TYPE="msys" ;;
  *cygwin*)     OS_TYPE="cygwin" ;;
  *)            OS_TYPE="unknown" ;;
esac

# Point this to the exact package path where your main function lives
PKG="main"

echo "Building bedrock_helper-$OS_TYPE-$CPU_ARCHITECTURE version $VERSION..."

# Run build from the root, pointing to the main package directory
go build -ldflags "-X '${PKG}.Version=${VERSION}' -X '${PKG}.CommitHash=${COMMIT}' -X '${PKG}.BuildDate=${DATE}'" -o $PROJECTROOT/bin/bedrock_helper-$OS_TYPE-$CPU_ARCHITECTURE ./cmd/bedrock_helper

echo "Build complete!"
