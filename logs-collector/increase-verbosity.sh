#!/bin/bash

hco_name=$(kubectl get -n $HCO_NAMESPACE hco -o json | jq -r .items[0].metadata.name)
if [[ -n "$hco_name" ]] ;
then
  kubectl patch -n $HCO_NAMESPACE hco $hco_name -p='[{"op": "add", "path": "/spec/configuration/logVerbosityConfig/kubevirt", "value":{"virtLauncher":4,"virtHandler":4}}]' --type='json'
  exit 0
fi

kubevirt_name=$(kubectl get -n $namespace kubevirt -o json | jq -r .items[0].metadata.name)
kubectl patch -n $namespace kubevirt $kubevirt_name -p='[{"op": "add", "path": "/spec/configuration/developerConfiguration/logVerbosity", "value":{"virtLauncher":4,"virtHandler":4}}]' --type='json'



