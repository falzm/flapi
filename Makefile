all: flapi

flapi:
	@gb build cmd/flapi

clean:
	@rm -rf bin/ pkg/
