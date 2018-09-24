#!/bin/bash

set -e -x -u

go fmt ./cmd/... ./pkg/... ./test/...

# export GOOS=linux GOARCH=amd64
go build ./cmd/...

./kwt version

./hack/generate-docs.sh

echo "Success"
