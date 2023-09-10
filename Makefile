TAG ?= 0.0.1

build-docker:
		docker build -t github.com/jscottransom/distributed_godis:$(TAG) .