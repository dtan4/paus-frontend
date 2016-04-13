default: build

build: clean
	go build -o bin/paus-frontend

clean:
	rm -f bin/paus-frontend

deps:
	go get -u github.com/Masterminds/glide
	glide install

docker-build: clean
	docker build -t quay.io/dtan4/paus-frontend:latest .

.PHONY: build clean docker-build
