#!/bin/bash
NAME=$1

kind get clusters | grep "${NAME}" > /dev/null
if [ $? -eq 0 ]; then
    kind delete cluster --name "${NAME}"
else
    echo "Cluster \"${NAME}\" does not exist."
fi