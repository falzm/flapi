all: flapi

flapi:
	@gb build flapi

clean:
	@rm -rf bin/ pkg/
