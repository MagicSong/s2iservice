#!/bin/bash
#运行主程序main

set -o errexit
set -o nounset
set -o pipefail

ROOT=$(dirname $BASH_SOURCE)/..
echo $ROOT
cd $ROOT

#set enviroment varibles for test
export DEVOPSPHERE_MYSQL_HOST=192.168.98.3
export DEVOPSPHERE_MYSQL_PORT=3306 
export DEVOPSPHERE_MYSQL_USER=root
export DEVOPSPHERE_MYSQL_PASSWORD=welcometodevops
export DEVOPSPHERE_MYSQL_DATABASE=TestDB

go build -v  -o cmd/server build/main.go
cmd/server
