#!/bin/bash
NAME=${1:-kind}
KUBECONFIG=/tmp/admin.conf

kind get clusters | grep "${NAME}" > /dev/null
if [ $? -eq 1 ]; then
    kind create cluster --name "${NAME}"
else
    echo "Cluster \"${NAME}\" already exists."
fi

docker pull kubernetes/pause
kind load docker-image --name "${NAME}" kubernetes/pause
kind get kubeconfig --name "${NAME}" > ${KUBECONFIG}

echo "Output kubeconfig to ${KUBECONFIG}"