VERSION := 0.1.0
BUILD_DATE := $(shell date +%F)
DOCKER_TAG ?= flapi:$(VERSION)

all: flapi flapctl

flapi:
	@gb build -ldflags "\
		-X main.version=$(VERSION) \
		-X main.buildDate=$(BUILD_DATE) \
		" \
		cmd/flapi

flapctl:
	@gb build cmd/flapctl

docker:
	@docker build \
			--build-arg VERSION=$(VERSION) \
			--build-arg BUILD_DATE=$(BUILD_DATE) \
			-t $(DOCKER_TAG) .

clean:
	@rm -rf bin/ pkg/
