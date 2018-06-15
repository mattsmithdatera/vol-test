# Makefile for the Docker image
# Authors: Keith Hudgins, Matt Smith

.PHONY: build test container

PREFIX ?= docker
TAG ?= v0.1.0
NODE1 ?= dockeree2-1
NODE2 ?= dockeree2-2

build:
	go get -d ./...
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bin/vol-test cmd/vol-test/main.go

build-local:
	docker-machine env $(NODE1) > node1
	docker-machine env $(NODE2) > node2
	go get -d ./...
	GOARCH=amd64 CGO_ENABLED=0 go build -o bin/vol-test cmd/vol-test/main.go

clean:
	rm -f bin/vol-test

container: build
	docker-machine env $(NODE1) > node1
	docker-machine env $(NODE2) > node2
	mkdir -p .docker/machine/machines
	cp -r $(HOME)/.docker/machine/machines/$(NODE1) .docker/machine/machines
	cp -r $(HOME)/.docker/machine/machines/$(NODE2) .docker/machine/machines
	docker build -t $(PREFIX)/vol-test:$(TAG) --build-arg home=$(HOME) --build-arg node1=$(NODE1) .
	docker build -t $(PREFIX)/vol-test:latest --build-arg home=$(HOME) --build-arg node2=$(NODE2) .
	rm -f -- node1
	rm -f -- node2
	rm -rf -- .docker

test:
	go test $$(go list ./... | grep -v /vendor/)
