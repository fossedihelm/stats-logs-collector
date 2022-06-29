# Stats-log-collector

**Stats-log-collector** is a deployment that could be used to monitor RSS 
and collect logs of the virtual machine inside a cluster [Kubernetes][k8s] that uses
[Kubevirt][kubevirt].

## Components

**Stats-log-collector** deployment consist in 2 containers:
- logs-collector
- memstat
Both uses a volume that is mounted, in which they will store data collected. By default,
the volume is mounted on `/data`

### logs-collector

logs-collector collects _virt-launcher_ and _virt-handler_ logs of all the vms running
in a specified _namespace_. It consists in a basic application that runs every hour
and retrieve the _virt-handler_ and _virt-launcher_ logs of the last hour. It will create
a folder inside `DATA_DIR` for each vm.

You can customize the behaviour by editing the following `env` of the specific container
in the [deployment.yaml](./deployment.yaml) file:
- `DATA_DIR` folder where store the collected logs
  - default: `/logs-collector`
- `KUBEVIRT_NAMESPACE` namespace in which [Kubevirt][kubevirt] has been deployed
  - default: `kubevirt`
- `HCO_NAMESPACE` namespace in which [HCO][hco] has been deployed (if exists)
   - default: ""
- `NAMESPACE` _namespace_ of the vms you want to monitor (you can monitor only one _namespace_)
   - default: `default`


### memstats

memstats collects RSS data of all the vms running in a specified _namespace_.
It consists in a basic application that runs every `POLL_INTERVAL_SECS` and 
get RSS data, adding them to a _csv_ file. This file will be stored inside a
directory named `mem-stats.csv` under `DATA_DIR` folder.

You can customize the behaviour by editing the following `env` of the specific container
in the [deployment.yaml](./deployment.yaml) file:
- `DATA_DIR` folder where store the collected logs
   - default: `/data`
- `NAMESPACE` _namespace_ of the vms you want to monitor (you can monitor only one _namespace_)
   - default: `default`
- `POLL_INTERVAL_SECS` every how many seconds you want the data collection to run
   - default: `5`
- `HTTP_PORT` port of the webserver that can be used to retrieve the _csv_ file
  - default: `8099`

### Usage

NB: it uses the `KUBECONFIG` env var on your local system to determine in which cluster
the commands should be executed.

After customizing the `env` variables you can install the **stats-logs-collector**
by simply run:
```shell
make install 
```

To uninstall it:
```shell
make unistall
```
It will destroy everything, also the **PVC** containing the collected data.


## Export data

### Export `mem-stats.csv` data

**memstat** provides a webserver that can be used to download the _csv_.

In order to use it you have to provide a port-forward to map the exposed port.

1. open a port forward to an available local port
```shell
kubectl port-forward <pod_name>  <local_port>>:<HTTP_PORT>
```

2. open another shell and retrieve data
```shell
curl -k http://localhost:<local_port> > <path_to_local_file>
```


### Read logs-collector data ### 

To read the logs you can simply go inside the pod and read it navigating through `DATA_DIR` folder:
```shell
kubectl exec -it <pod_name>  -- bash
```

or download a specific file:
```shell
kubectl cp -c logs-collector <pod_name>:<DATA_DIR>/<vm_name>/<filename> <local_path>
```

## License

Stats-logs-collector is distributed under the
[Apache License, Version 2.0](http://www.apache.org/licenses/LICENSE-2.0.txt).

    Copyright 2022

    Licensed under the Apache License, Version 2.0 (the "License");
    you may not use this file except in compliance with the License.
    You may obtain a copy of the License at

        http://www.apache.org/licenses/LICENSE-2.0

    Unless required by applicable law or agreed to in writing, software
    distributed under the License is distributed on an "AS IS" BASIS,
    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
    See the License for the specific language governing permissions and
    limitations under the License.

[//]: # (Reference links)
   [k8s]: https://kubernetes.io
   [kubevirt]: https://kubevirt.io
   [hco]: https://github.com/kubevirt/hyperconverged-cluster-operator
   [crd]: https://kubernetes.io/docs/tasks/access-kubernetes-api/extend-api-custom-resource-definitions/
   [ovirt]: https://www.ovirt.org
   [cockpit]: https://cockpit-project.org/
   [libvirt]: https://www.libvirt.org
   [kubevirt-ansible]: https://github.com/kubevirt/kubevirt-ansible
