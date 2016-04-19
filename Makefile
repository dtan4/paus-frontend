default: build

build: clean
	go build -o bin/paus-frontend

build-linux: clean
	GOOS=linux GOARCH=amd64 go build -o bin/paus-frontend-linux_amd64

clean:
	rm -f bin/paus-frontend
	rm -f bin/paus-frontend-linux_amd64

deps:
	go get -u github.com/Masterminds/glide
	glide install

docker-build: clean
	docker build -t quay.io/dtan4/paus-frontend:latest .

docker-build-release: build-linux
	docker build -f Dockerfile.release -t quay.io/dtan4/paus-frontend:latest .

.PHONY: build clean docker-build
