TAG ?= 0.0.1
CONFIG_PATH=${HOME}/distributed_projects/.distributed_godis/

build-docker:
		docker build -t github.com/jscottransom/distributed_godis:$(TAG) .


.PHONY: init
init:
	mkdir -p ${CONFIG_PATH}

.PHONY: gencert
gencert:
		cfssl gencert \
				-initca auth/ca-csr.json | cfssljson -bare ca
		
		cfssl gencert \
				-ca=ca.pem \
				-ca-key=ca-key.pem \
				-config=auth/ca-config.json \
				-profile=server
				auth/server-csr.json | cfssljson -bare server
				mv *.pem *.csr ${CONFIG_PATH}
