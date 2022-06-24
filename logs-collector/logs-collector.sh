#!/bin/bash


if [ "$#" -ne 6 ]
then
	echo "Usage: mem-collector.sh -n <namespace> -d <output_directory> -kn <kubevirt_namespace>"
	exit 1
fi

while (( $# )); do
	case $1 in
	-n|--namespace) namespace=$2
			shift
			;;
	-kn|--kv_namespace) kubevirt_namespace=$2
			shift
			;;
	-d|--directory) directory=$2
			shift
			;;
	*) echo "Usage: mem-collector.sh -n <namespace> -d <output_directory> -kn <kubevirt_namespace>"
			exit 1
	esac
	shift
done
date=$(date +"%m.%d_%Hh%Mm")
local_port=6560
vms=$(kubectl get -n $namespace vm -o json | jq -r .items[].metadata.name)
vms=($vms)
for vm_name in "${vms[@]}"
do
	mkdir -p $directory/$vm_name
	cd $directory/$vm_name


	echo -e "\n---== $date ==---\n"

	kubectl get vm $vm_name
	launcher_pod=$( kubectl get pod -l vm.kubevirt.io/name=$vm_name -o json | jq -r .items[].metadata.name)
	kubectl get pod $launcher_pod

	# get launcher logs
	kubectl logs --since=1h $launcher_pod > $launcher_pod.$date.log

	# get handler logs
	vmi_node=$(kubectl get vmi $vm_name -o json | jq -r .status.nodeName)
	handler_pod=$(kubectl get pod -n $kubevirt_namespace --field-selector spec.nodeName=$vmi_node -l kubevirt.io=virt-handler -o json | jq -r .items[0].metadata.name)
	kubectl logs -n $kubevirt_namespace $handler_pod --since=1h > $handler_pod.$date.log

	# start port forward
	echo Enabling port-forward from local:$local_port to $launmcher_pod:6060
	kubectl port-forward $launcher_pod $local_port:6060 &
	pid=$!

  sleep 2
#	# wait for $localport to become available
#	while ! nc -vz localhost $local_port > /dev/null 2>&1 ; do
#    	# echo sleeping
#    	sleep 0.1
#	done

	# get heap profile
	curl http://localhost:$local_port/debug/pprof/heap > heap.$date.pprof

	# kill port forward
	kill $pid

  local_port=$(expr $local_port + 1)
done


