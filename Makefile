default: build

all: build push uninstall install

build:
	REGISTRY=${REGISTRY:-quay.io/acardace}
	cd logs-collector && docker build -t ${REGISTRY}/logs-collector:latest .
	cd memstat && docker build -t ${REGISTRY}/memstat:latest .

push: build
	REGISTRY=${REGISTRY:-quay.io/acardace}
	docker push ${REGISTRY}/logs-collector:latest
	docker push ${REGISTRY}/memstat:latest

install: generate
	-kubectl create -f ./logs-collector/rbac.yaml
	-kubectl create -f ./memstat/rbac.yaml
	-kubectl create -f ./_out/deployment_generated.yaml

uninstall: generate
	-kubectl delete -f ./logs-collector/rbac.yaml
	-kubectl delete -f ./memstat/rbac.yaml
	-kubectl delete -f ./_out/deployment_generated.yaml

generate:
	REGISTRY=${REGISTRY:-quay.io/acardace}
	rm _out/* -rf
	sed "s#<REGISTRY>#${REGISTRY}#" deployment.yaml > _out/deployment_generated.yaml
