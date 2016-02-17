default: build

build: clean
	go build -o bin/paus-frontend

clean:
	rm -f bin/paus-frontend

docker-build: clean
	docker build -f Dockerfile.build -t quay.io/dtan4/paus-frontend:latest .

.PHONY: build clean docker-build
