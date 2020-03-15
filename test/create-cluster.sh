#!/bin/bash
NAME=${1:-kind}
KUBECONFIG=/tmp/admin.conf

load_image() {
    docker pull $1
    kind load docker-image --name "${NAME}" $1
}

kind get clusters | grep "${NAME}" > /dev/null
if [ $? -eq 1 ]; then
    kind create cluster --name "${NAME}"
else
    echo "Cluster \"${NAME}\" already exists."
fi

load_image kubernetes/pause
load_image alpine
kind get kubeconfig --name "${NAME}" > ${KUBECONFIG}

echo "Output kubeconfig to ${KUBECONFIG}"