#!/bin/bash

set -e -x -u

ytt -f images/sshd/ | kbld -f- | ytt -f image.yml=- -f pkg/kwt/cmd/net/ssh_flags_generated.go.txt --output-directory ./tmp/

mv ./tmp/ssh_flags_generated.go.txt pkg/kwt/cmd/net/ssh_flags_generated.go
