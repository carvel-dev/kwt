#!/bin/bash

set -e -x -u

export KWT_KUBECONFIG=
export KWT_NAMESPACE=

go run ./hack/generate-docs.go

echo "Success"
