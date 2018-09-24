#!/bin/bash

set -e -x -u

apt-get -y update
apt-get -y install wget curl

wget -O- https://dl.google.com/go/go1.10.3.linux-amd64.tar.gz > /tmp/go
echo "fa1b0e45d3b647c252f51f5e1204aba049cde4af177ef9f2181f43004f901035  /tmp/go" | sha256sum -c
tar -C /usr/local -xzf /tmp/go
export PATH=$PATH:/usr/local/go/bin

export GOPATH=$PWD/gopath
cd $GOPATH/src/github.com/cppforlife/kwt

./hack/build.sh
./hack/test.sh
