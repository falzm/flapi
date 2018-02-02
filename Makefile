all: flapi flapctl

flapi:
	@gb build cmd/flapi

flapctl:
	@gb build cmd/flapctl

clean:
	@rm -rf bin/ pkg/
