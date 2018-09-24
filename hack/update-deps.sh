#!/bin/bash

set -e -x -u

dep ensure

rm -rf $(find vendor/ -name 'OWNERS')
rm -rf $(find vendor/ -name '*_test.go')

# TODO update licenses
