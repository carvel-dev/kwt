#!/bin/bash

set -e -x -u

apt-get -y update
apt-get -y install wget curl gnupg iptables

apt-get install -y apt-transport-https
curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key add -
echo "deb http://apt.kubernetes.io/ kubernetes-xenial main" > /etc/apt/sources.list.d/kubernetes.list
apt-get -y update
apt-get -y install kubectl

wget -O- https://dl.google.com/go/go1.10.3.linux-amd64.tar.gz > /tmp/go
echo "fa1b0e45d3b647c252f51f5e1204aba049cde4af177ef9f2181f43004f901035  /tmp/go" | sha256sum -c
tar -C /usr/local -xzf /tmp/go
export PATH=$PATH:/usr/local/go/bin

export GOPATH=$PWD/gopath
cd $GOPATH/src/github.com/carvel-dev/kwt

./hack/build.sh
ln -sf $PWD/kwt /usr/local/bin/kwt

export KWT_NAMESPACE=kwt-$(date +%s%N | sha256sum | cut -f1 -d' ' | head -c 32)
kwt resource create-ns
function finish {
  kubectl delete ns $KWT_E2E_NAMESPACE
}
trap finish EXIT

export KWT_E2E_NAMESPACE=$KWT_NAMESPACE
./hack/test-e2e.sh
