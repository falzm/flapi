VERSION := 0.1.0dev
BUILD_DATE := $(shell date +%F)
DOCKER_TAG ?= flapi:$(VERSION)

all: flapi

flapi:
	@GOPATH="$(PWD):$(PWD)/vendor" \
		go build -ldflags "\
			-X main.version=$(VERSION) \
			-X main.buildDate=$(BUILD_DATE) \
			" \
			./src/cmd/flapi

docker:
	@docker build \
			--build-arg VERSION=$(VERSION) \
			--build-arg BUILD_DATE=$(BUILD_DATE) \
			-t $(DOCKER_TAG) .

chaos:
	@curl -X PUT -H 'Content-Type: application/json' \
		-d '{"error":{"status_code":504,"p":1},"delay":{"duration":3000,"p":0.5}}' \
		'localhost:8666/?method=POST&path=/api/a'
	@curl 'localhost:8666/?method=POST&path=/api/a'
	@curl -X PUT -H 'Content-Type: application/json' \
		-d '{"error":{"status_code":599,"message":"lolnope!","p":1}}' \
		'localhost:8666/?method=GET&path=/api/b'
	@curl 'localhost:8666/?method=GET&path=/api/b'

clean:
	@rm -rf flapi bin/ pkg/
