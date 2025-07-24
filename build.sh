#!/usr/bin/env bash

set -e

export CC=gcc
export CGO_ENABLED=1
export GOPATH="${PWD}/go"
export GOROOT="${PWD}/go-sdk"
export GOARCH arm
export OOS=linux
export PATH="${PATH}:${PWD}/go-sdk/bin:${PWD}/go/bin"

echo "*** Generate version file ***"
go generate pkg/server/version.go

echo "*** Compile binary ***"
go build -tags static -trimpath --ldflags '-s -w' -o elrs-joystick-control ./cmd/elrs-joystick-control/.

echo "*** Create distribution zip file ***"
go run scripts/cmd/build-release-zip/build-release-zip.go --location . --prefix elrs-joystick-control --files *-control,LICENSE*
