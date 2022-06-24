#!/bin/bash

if [ "$#" -ne 2 ]
then
	echo "Usage: increase-verbosity-kv.sh -kn <kubevirt_namespace>"
	exit 1
fi

while (( $# )); do
	case $1 in
	-n|--namespace) namespace=$2
			shift
			;;
	*) echo "Usage: increase-verbosity-kv.sh -n <namespace>"
			exit 1
	esac
	shift
done

hco_name=$(kubectl get -n $namespace hco -o json | jq -r .items[0].metadata.name)
if [[ -n "$hco_name" ]] ;
then
  kubectl patch -n $namespace hco $hco_name -p='[{"op": "add", "path": "/spec/configuration/logVerbosityConfig/kubevirt", "value":{"virtLauncher":4,"virtHandler":4}}]' --type='json'
  exit 0
fi

kubevirt_name=$(kubectl get -n $namespace kubevirt -o json | jq -r .items[0].metadata.name)
kubectl patch -n $namespace kubevirt $kubevirt_name -p='[{"op": "add", "path": "/spec/configuration/developerConfiguration/logVerbosity", "value":{"virtLauncher":4,"virtHandler":4}}]' --type='json'



