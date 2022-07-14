#!/bin/bash
KIND_IMAGE=kindest/node:v1.23.0
set -x
kind create cluster --name px-object-controller --config hack/kind.yaml --image $KIND_IMAGE
kind get kubeconfig --name px-object-controller > /tmp/px-object-controller-kubeconfig.yaml 
