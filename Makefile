#
# Copyright (c) 2018 Tencent
#
# SPDX-License-Identifier: Apache-2.0
#

.PHONY: build clean test run docker run_docker

GO=CGO_ENABLED=0 GO111MODULE=on go
GOCGO=CGO_ENABLED=1 GO111MODULE=on go

MICROSERVICES=cmd/object-manager
.PHONY: $(MICROSERVICES)

DOCKERS=docker_object_manager
.PHONY: $(DOCKERS)

VERSION=$(shell cat ./VERSION)

GOFLAGS=-ldflags "-X github.com/object-manager.Version=$(VERSION)"

GIT_SHA=$(shell git rev-parse HEAD)

build: $(MICROSERVICES)
	$(GO) build ./...

cmd/object-manager:
	$(GO) build $(GOFLAGS) -o $@ ./cmd

clean:
	rm -f $(MICROSERVICES)

test:
	GO111MODULE=on go test -coverprofile=coverage.out ./...
	GO111MODULE=on go vet ./...

prepare:

run:
	cd bin && ./object-manager-launch.sh

run_docker: 
	cd cmd && ./object-manager -r=true -p=docker
docker: $(DOCKERS)

docker_object_manager:
		docker build \
		--label "git_sha=$(GIT_SHA)" \
		-t phanvanhai/docker-object-manager:$(VERSION) \
		.
