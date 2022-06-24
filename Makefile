default: build

all: build push install

build:
	cd logs-collector && docker build -t quay.io/acardace/logs-collector:latest .
	cd memstat && docker build -t quay.io/acardace/memstat:latest .

push:
	docker push quay.io/acardace/logs-collector:latest
	docker push quay.io/acardace/memstat:latest

install:
	-kubectl create -f ./logs-collector/logs-collector-deployment.yaml
	-kubectl create -f ./memstat/deploy.yaml


uninstall:
	-kubectl delete -f ./logs-collector/logs-collector-deployment.yaml
	-kubectl delete -f ./memstat/deploy.yaml

