#!/bin/bash
#****************************************************************************
# script to start NetApp python API for go-plugin
# NOTE: requires virtualenv installed
# NOTE: must be executed with API root folder followed by script as arguments
#        with additional arguments following spaced by space :)
#	./scripts/start_api.sh $PYTHON_API_ROOT$ $API_SCRIPT$ ...
#****************************************************************************

APIROOT="$1"
shift

# navigate to Python API root directory
cd $APIROOT

echo "Starting NetApp Python API in: ${APIROOT}" >> python_api.log

# check if virtualenv is already present, create if not
if [ ! -d "./venv" ]
then
   echo "ERROR: Python API virtualenv not installed!"
   exit 1   # flag error?
fi

# activate virtualenv
source venv/bin/activate
# start python API
nohup python "$@" &
# save PID to file
echo $! > python_api.pid