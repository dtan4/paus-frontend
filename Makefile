BINARY := paus-frontend
BINARY_DIR := bin
DOCKER_REPOSITORY := quay.io
DOCKER_IMAGE_NAME := $(DOCKER_REPOSITORY)/dtan4/paus-frontend
DOCKER_IMAGE_TAG := latest
DOCKER_IMAGE := $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)

default: build

build: clean
	go build -o $(BINARY_DIR)/$(BINARY)

build-linux: clean
	GOOS=linux GOARCH=amd64 go build -o $(BINARY_DIR)/$(BINARY)-linux_amd64

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

test:
	go test

.PHONY: build build-linux ci-docker-release clean deps docker-build docker-release-build test
