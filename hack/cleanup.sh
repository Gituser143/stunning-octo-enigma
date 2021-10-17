#!/bin/bash

function clean_evicted {
    echo "cleaning up evicted pods"
    kubectl get pods --all-namespaces | grep Evicted | awk '{print $2 " --namespace=" $1}' | xargs kubectl delete pod
}

function scale_to_one {
    echo "scaling back to 1"
    kubectl get deployments -n istio-teastore | awk '{print $1}' | tail -n +2 | xargs kubectl scale --replicas=1 deployment -n istio-teastore
}

if [[ $# -lt 1 ]]; then
    echo "missing args: need [evicted | scale | all]"
    exit 1
fi

if [[ $1 == "all" ]]; then
    clean_evicted
    scale_to_one
fi

if [[ $1 == "evicted" ]]; then
    clean_evicted
fi

if [[ $1 == "scale" ]]; then
    scale_to_one
fi
