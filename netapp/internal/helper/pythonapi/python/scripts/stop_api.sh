#!/bin/bash
#****************************************************************************
# script to 'roughly' stop/kill NetApp python API for go-plugin
# NOTE: requires virtualenv installed
# NOTE: must be executed with API root folder as argument
#	./scripts/stop_api.sh $PYTHON_API_ROOT$
#****************************************************************************

APIROOT="$1"
# navigate to Python API root directory
cd $APIROOT

echo "Stopping NetApp Python API in: ${APIROOT}"

pid_file="./python_api.pid"
api_sd_file="./shut_api"


# create shutdown file and wait for 10s
touch $api_sd_file
sleep 10

# check if python api PID file exists
if [[ -f $pid_file ]]
then
   echo "NetApp Python API PID file exists, killing!"
   local pid=$(< "$pid_file")
   kill pid
   rm -vf $pid_file
   exit 0
else
   echo "ERROR: no NetApp Python API PID file <-- not started?"
   exit 1
fi