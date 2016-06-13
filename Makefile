VERSION := 0.1.0
REVISION := $(shell git rev-parse --short HEAD)
BUILDTIME := $(shell date '+%Y/%m/%d %H:%M:%S %Z')

BINARY := paus-frontend
BINARY_DIR := bin

LDFLAGS := -ldflags="-w -X \"main.Version=$(VERSION)\" -X \"main.Revision=$(REVISION)\" -X \"main.BuildTime=$(BUILDTIME)\""

ETCD_CONTAINER := etcd

DOCKER_REPOSITORY := quay.io
DOCKER_IMAGE_NAME := $(DOCKER_REPOSITORY)/dtan4/paus-frontend
DOCKER_IMAGE_TAG := latest
DOCKER_IMAGE := $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)

default: build

build: clean
	go build $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY)

build-linux: clean
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY)_linux-amd64

ci-docker-release: docker-release-build
	@docker login -e="$(DOCKER_QUAY_EMAIL)" -u="$(DOCKER_QUAY_USERNAME)" -p="$(DOCKER_QUAY_PASSWORD)" $(DOCKER_REPOSITORY)
	docker push $(DOCKER_IMAGE)

clean:
	rm -fr $(BINARY_DIR)

deps:
	go get -u github.com/Masterminds/glide
	glide install

docker-build: clean
	docker build -t $(DOCKER_IMAGE) .

docker-release-build: build-linux
	docker build -f Dockerfile.release -t $(DOCKER_IMAGE) .

run-etcd: stop-etcd
	docker run -d -p 4001:4001 -p 2380:2380 -p 2379:2379 --name $(ETCD_CONTAINER) \
		quay.io/coreos/etcd:v2.3.6 \
			-name etcd0 \
			-advertise-client-urls http://0.0.0.0:2379,http://0.0.0.0:4001 \
			-listen-client-urls http://0.0.0.0:2379,http://0.0.0.0:4001 \
			-initial-advertise-peer-urls http://0.0.0.0:2380 \
			-listen-peer-urls http://0.0.0.0:2380 \
			-initial-cluster-token etcd-cluster-1 \
			-initial-cluster etcd0=http://0.0.0.0:2380 \
			-initial-cluster-state new

stop-etcd:
	docker stop $(ETCD_CONTAINER) > /dev/null 2>&1 || true
	docker rm $(ETCD_CONTAINER) > /dev/null 2>&1 || true

test:
	go test

.PHONY: build build-linux ci-docker-release clean deps docker-build docker-release-build test
