#!/bin/bash
#编译主程序

set -o errexit
set -o nounset
set -o pipefail

ROOT=$(dirname $BASH_SOURCE)/..
echo $ROOT
cd $ROOT

CGO_ENABLED=0 GOOS=linux go build -v  -a -installsuffix cgo -ldflags '-w'  -o cmd/server build/main.go

docker build -f build/dockerfile -t s2i-builder:0.1 cmd/
