#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

# variables
TAG=s2i-builder:test
CONTAINER_NAME=s2itest
RUNHOST=localhost

mkdir -p /tmp/s2iservice
WORK_DIR=$(mktemp -d /tmp/s2iservice/test-work.XXXX)
# functions
function cleanup(){
   set +e
   echo "---Begin to cleanup----"
   local result=$(docker inspect -f '{{.State.Running}}' $CONTAINER_NAME)
   if [ $? -eq 0 ]; then 
    if [ $result = "true" ]; then
        echo "Stop running container"
        docker container stop $CONTAINER_NAME 2&>1 >/dev/null
        echo "Container stopped successfully"    
    fi 
   fi
   rm -rf $WORK_DIR
   echo "---Cleanup done----"
}

## $1 is exit code, $2 is the path of logfile
function check_result() {
    local result=$1
    if [ $result -eq 0 ]; then
        echo "TEST PASSED"
        echo
        return
    fi
    echo
    echo "TEST FAILED ${result}"
    echo
    cat $2
    cleanup
    exit $result
}

function test_debug() {
    echo
    echo $1
    echo
}

trap cleanup EXIT SIGINT
ROOT=$(dirname $BASH_SOURCE)/..
cd $ROOT

## Get Docker Host
if [ -n ${DOCKER_HOST:-} ]; then
    RUNHOST=$(echo "$DOCKER_HOST" | awk '{match($0,/[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+/); ip = substr($0,RSTART,RLENGTH); print ip}')
    if [ -z $RUNHOST ]; then
        echo "DOCKER_HOST is set correctly,the <$DOCKER_HOST> doesn't have an ip address"
        exit 1
    fi
fi
## Build Image

hack/build-image.sh $TAG

## Run Container
CONTAINER_ID=$(docker run --name $CONTAINER_NAME -d --rm -p 8001:8001 $TAG)

timeout=0
while [ $timeout -lt 15 ]
do
    docker logs $CONTAINER_NAME | grep "s2i start to watch tasks queue"
    if [ $? -eq 0 ]; then
        echo "detect that container is running successfully"
        break
    fi
    let timeout=$timeout+1
    sleep 1
done

if [ $timeout -eq 15 ]; then
    echo "Container start FAILED Due to FAILED, Please check the following logs"
    docker logs $CONTAINER_NAME
    exit 1
fi

##test begin

test_debug "API TEST- Get Templates"
logfile=$WORK_DIR/test.log
curl --request GET -I  --silent --url http://$RUNHOST:8001/api/v1alpha1/templates --header 'authorization: Basic YWRtaW46YWRtaW5z' > $logfile
cat $logfile | grep "HTTP/1.1 200" >/dev/null
check_result $? $logfile
