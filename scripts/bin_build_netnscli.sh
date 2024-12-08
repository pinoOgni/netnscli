#!/bin/bash

SCRIPTS_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)
PROJECT_DIR=$SCRIPTS_DIR/..
BIN_DIR=$PROJECT_DIR/bin
CMD_DIR=$PROJECT_DIR/cmd

GO=${GO:-go}
GOOS=${GOOS:-linux}
GOARCH=${GOARCH:-amd64}

set -x -e

# TODO: replace hardcoded ldflags with values coming from env
eval "CGO_ENABLED=0 GOOS=$GOOS GOARCH=$GOARCH \
  $GO build -ldflags='-s -w' $GO_BUILD_FLAGS -o $BIN_DIR/$BIN $CMD_DIR/main.go"

