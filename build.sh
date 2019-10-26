#!/bin/bash

scriptDir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

docker build ${scriptDir} -f ${scriptDir}/docker/Dockerfile -t iotd --network host
