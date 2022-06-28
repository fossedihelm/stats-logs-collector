default: build

all: build push uninstall install

build:
	cd logs-collector && docker build -t quay.io/acardace/logs-collector:latest .
	cd memstat && docker build -t quay.io/acardace/memstat:latest .

push: build
	docker push quay.io/acardace/logs-collector:latest
	docker push quay.io/acardace/memstat:latest

install: push
	-kubectl create -f ./logs-collector/rbac.yaml
	-kubectl create -f ./memstat/rbac.yaml
	-kubectl create -f ./deployment.yaml


uninstall:
	-kubectl delete -f ./logs-collector/rbac.yaml
	-kubectl delete -f ./memstat/rbac.yaml
	-kubectl delete -f ./deployment.yaml
