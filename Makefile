TAG ?= 0.0.1
CONFIG_PATH=${HOME}/dev/.distributed_godis/

build-docker:
		docker build -t github.com/jscottransom/distributed_godis:$(TAG) .


.PHONY: init
init:
	mkdir -p ${CONFIG_PATH}

.PHONY: gencert
gencert:
		cfssl gencert \
				-initca test/ca-csr.json | cfssljson -bare ca
		
		cfssl gencert \
				-ca=ca.pem \
				-ca-key=ca-key.pem \
				-config=test/ca-config.json \
				-profile=server \
				test/server-csr.json | cfssljson -bare server

		cfssl gencert \
				-ca=ca.pem \
				-ca-key=ca-key.pem \
				-config=test/ca-config.json \
				-profile=client \
				-cn="root" \
				test/client-csr.json | cfssljson -bare root-client
		
		cfssl gencert \
				-ca=ca.pem \
				-ca-key=ca-key.pem \
				-config=test/ca-config.json \
				-profile=client \
				-cn="nobody" \
				test/client-csr.json | cfssljson -bare nobody-client
		mv *.pem *.csr ${CONFIG_PATH}


$(CONFIG_PATH)/model.conf:
	cp test/model.conf $(CONFIG_PATH)/model.conf

$(CONFIG_PATH)/policy.csv:
	cp test/policy.csv $(CONFIG_PATH)/policy.csv

.PHONY: test
test: $(CONFIG_PATH)/policy.csv $(CONFIG_PATH)/model.conf
		go test -race -v ./...
