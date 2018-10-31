FROM golang:1 AS builder
ARG VERSION
ARG BUILD_DATE
COPY . /flapi
WORKDIR /flapi
RUN make GOOS=linux CGO_ENABLED=0 flapi

#--------------------------------------------------------------------

FROM scratch
COPY --from=builder /flapi/flapi /usr/bin/flapi
COPY flapi.yaml /etc/
ENTRYPOINT ["/usr/bin/flapi"]  
CMD ["-config", "/etc/flapi.yaml"]
