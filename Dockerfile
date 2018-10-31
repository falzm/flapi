FROM golang:1 AS builder
ARG VERSION
ARG BUILD_DATE
COPY . /flapi
WORKDIR /flapi
RUN GOPATH=/flapi:/flapi/vendor GOOS=linux CGO_ENABLED=0 go build \
        -a -installsuffix nocgo \
        -ldflags "\
		        -X main.version=${VERSION} \
		        -X main.buildDate=${BUILD_DATE} \
		" \
        ./src/cmd/flapi

#--------------------------------------------------------------------

FROM scratch
COPY --from=builder /flapi/flapi /usr/bin/flapi
COPY flapi.yaml /etc/
ENTRYPOINT ["/usr/bin/flapi"]  
CMD ["-config", "/etc/flapi.yaml"]
