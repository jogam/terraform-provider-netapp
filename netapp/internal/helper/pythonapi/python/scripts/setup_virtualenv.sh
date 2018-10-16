#!/bin/bash
#****************************************************************************
# script to install virtualenv and requirements for NetApp python API
# NOTE: requires virtualenv installed
# NOTE: must be executed with API root folder as argument
#	./scripts/setup_virtualenv.sh $PYTHON_API_ROOT$
#****************************************************************************

APIROOT="$1"
echo "Installing/Updating NetApp Python API virtualenv in: ${APIROOT}"

# navigate to Python API root directory
cd $APIROOT

# check if virtualenv is already present, create if not
if [ ! -d "./venv" ]
then
   echo "creating virtualenv for Python API"
   if command -v python3 &>/dev/null;
   then
      echo "Python3 exists using that..."
      virtualenv --python=python3 venv
   else
      echo "no Python3 using python..."
      virtualenv venv
   fi
fi

# activate virtualenv
source venv/bin/activate
# install pip requirements
pip install -r requirements.txt

# TODO: create some sort of python script to execute that verifies all is well?!

# deactivate virtualenv
deactivate