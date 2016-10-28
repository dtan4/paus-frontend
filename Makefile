NAME := paus-frontend
VERSION := 0.1.0
REVISION := $(shell git rev-parse --short HEAD)
GOVERSION := $(subst go version ,,$(shell go version))

SOURCES := $(shell find . -name '*.go' -type f)
LDFLAGS := -ldflags="-s -w -X \"main.Version=$(VERSION)\" -X \"main.Revision=$(REVISION)\" -X \"main.GoVersion=$(GOVERSION)\""

LINUX_AMD64_SUFFIX := _linux-amd64

GLIDE := $(shell command -v glide 2> /dev/null)

DOCKER_REPOSITORY := quay.io
DOCKER_IMAGE_NAME := $(DOCKER_REPOSITORY)/dtan4/paus-frontend
DOCKER_IMAGE_TAG := latest
DOCKER_IMAGE := $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)

.DEFAULT_GOAL := bin/$(NAME)

.PHONY: glide
glide:
ifndef GLIDE
	curl https://glide.sh/get | sh
endif

bin/$(NAME): deps $(SOURCES)
	go build $(LDFLAGS) -o bin/$(NAME)

bin/$(NAME)$(LINUX_AMD64_SUFFIX): deps $(SOURCES)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/$(NAME)$(LINUX_AMD64_SUFFIX)

.PHONY: ci-docker-release
ci-docker-release: docker-build
	@docker login -e="$(DOCKER_QUAY_EMAIL)" -u="$(DOCKER_QUAY_USERNAME)" -p="$(DOCKER_QUAY_PASSWORD)" $(DOCKER_REPOSITORY)
	docker push $(DOCKER_IMAGE)

.PHONY: clean
clean:
	rm -fr bin/*
	rm -fr vendor/*

.PHONY: deps
deps: glide
	glide install

.PHONY: docker-build
docker-build: bin/$(NAME)$(LINUX_AMD64_SUFFIX)
	docker build -t $(DOCKER_IMAGE) .

.PHONY: docker-push
docker-push:
	docker push $(DOCKER_IMAGE)

.PHONY: test
test: deps
	go test -v . ./model/app ./util
