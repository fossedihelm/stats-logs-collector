#!/bin/bash

set -e
date=$(date +"%m.%d_%Hh%Mm")
local_port=6560
vms=$(kubectl get -n $NAMESPACE vm -o json | jq -r .items[].metadata.name)
vms=($vms)
for vm_name in "${vms[@]}"
do
	mkdir -p $DATA_DIR/$vm_name
	cd $DATA_DIR/$vm_name

	echo -e "\n---== $date ==---\n"

	kubectl get vm $vm_name
	launcher_pod=$(kubectl get pod -l vm.kubevirt.io/name=$vm_name -o json | jq -r .items[].metadata.name)
	kubectl get pod $launcher_pod

	# get launcher logs
	kubectl logs --since=1h $launcher_pod > $launcher_pod.$date.log
  gzip -r $launcher_pod.$date.log

	# get handler logs
	vmi_node=$(kubectl get vmi $vm_name -o json | jq -r .status.nodeName)
	handler_pod=$(kubectl get pod -n $KUBEVIRT_NAMESPACE --field-selector spec.nodeName=$vmi_node -l kubevirt.io=virt-handler -o json | jq -r .items[0].metadata.name)
	kubectl logs -n $KUBEVIRT_NAMESPACE $handler_pod --since=1h > $handler_pod.$date.log
  gzip -r $handler_pod.$date.log

  if [[ "$PPROF_ENABLED" == "1" ]]
  then
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
    gzip -r heap.$date.pprof

    # kill port forward
    kill $pid

    local_port=$(expr $local_port + 1)
  fi
done


