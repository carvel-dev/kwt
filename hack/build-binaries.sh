#!/bin/bash

set -e -x -u

GOOS=darwin GOARCH=amd64 go build -o kwt-darwin-amd64 ./cmd/...
GOOS=linux GOARCH=amd64 go build -o kwt-linux-amd64 ./cmd/...

shasum -a 256 ./kwt-*-amd64
