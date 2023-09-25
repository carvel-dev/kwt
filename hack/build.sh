#!/bin/bash

set -e -x -u

VERSION="${1:-0.0.0+develop}"

export CGO_ENABLED=0
LDFLAGS="-X github.com/carvel-dev/kwt/pkg/kwt/cmd.Version=$VERSION"
repro_flags="-trimpath -mod=vendor"

go mod vendor
go mod tidy
go fmt ./cmd/... ./pkg/... ./test/...

# export GOOS=linux GOARCH=amd64
go build -ldflags="$LDFLAGS" $repro_flags ./cmd/...

./kwt version

./hack/generate-docs.sh

echo "Success"
