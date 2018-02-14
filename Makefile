VERSION := 0.1.0
BUILD_DATE := $(shell date +%F)

all: flapi flapctl

flapi:
	@gb	build -ldflags "\
		-X main.version=$(VERSION) \
		-X main.buildDate=$(BUILD_DATE) \
		" \
		cmd/flapi

flapctl:
	@gb build cmd/flapctl

clean:
	@rm -rf bin/ pkg/
