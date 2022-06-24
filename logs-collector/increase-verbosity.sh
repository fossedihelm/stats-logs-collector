#!/bin/bash

set -e
hco_name=$(kubectl get -n $HCO_NAMESPACE hco -o json | jq -r .items[0].metadata.name)
if [[ -n "$hco_name" ]] ;
then
  kubectl patch -n $HCO_NAMESPACE hco $hco_name -p='[{"op": "add", "path": "/spec/configuration/logVerbosityConfig/kubevirt", "value":{"virtLauncher":4,"virtHandler":4}}]' --type='json'
  exit 0
fi

kubevirt_name=$(kubectl get -n $KUBEVIRT_NAMESPACE kubevirt -o json | jq -r .items[0].metadata.name)
if [[ -n "$kubevirt_name" ]] ;
then
  kubectl patch -n $KUBEVIRT_NAMESPACE kubevirt $kubevirt_name -p='[{"op": "add", "path": "/spec/configuration/developerConfiguration/logVerbosity", "value":{"virtLauncher":4,"virtHandler":4}}]' --type='json'
  exit 0
fi
exit 1



