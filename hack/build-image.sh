#!/bin/bash
#编译主程序

set -o errexit
set -o nounset
set -o pipefail

ROOT=$(dirname $BASH_SOURCE)/..
cd $ROOT

CGO_ENABLED=0 GOOS=linux go build -v  -a -installsuffix cgo -ldflags '-w'  -o cmd/server build/main.go

if [ -z $1 ]; then
   echo "Please specify image tag"
   exit 1 
fi

docker build -f build/dockerfile -t $1 cmd/
