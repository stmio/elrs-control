#!/bin/bash

# build go demo
go build -tags static -trimpath --ldflags '-s -w' -o elrs-control-demo ./cmd/elrs-control/.

# Build shared library for Python
go build -buildmode=c-shared \
    -o elrs_control.so \
    ./cmd/pythonlib/
