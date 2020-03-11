.PHONY: test

VERSION?=$(shell git describe --tags)
IMAGE:=pod-reaper:$(VERSION)

CLUSTER_NAME=e2e 

all: build

build:
	CGO_ENABLED=0 go build -o _output/bin/pod-reaper github.com/target/pod-reaper/cmd/pod-reaper

image:
	docker build -t $(IMAGE) .

clean:
	rm -rf _output

test-unit:
	./test/run-unit-tests.sh

test-e2e:
	./test/create-cluster.sh $(CLUSTER_NAME)
	./test/run-e2e-tests.sh
	./test/delete-cluster.sh $(CLUSTER_NAME)