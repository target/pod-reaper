.PHONY: test

VERSION?=$(shell git describe --tags)
IMAGE:=pod-reaper:$(VERSION)

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
	kind create cluster
	docker pull kubernetes/pause
	kind load docker-image kubernetes/pause
	kind get kubeconfig > /tmp/admin.conf
	./test/run-e2e-tests.sh
	kind delete cluster