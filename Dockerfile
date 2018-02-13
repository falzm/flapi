FROM golang:1.9

COPY . /flapi

WORKDIR /flapi

RUN go get github.com/constabulary/gb/... && \
    make flapi && \
    mv bin/flapi /usr/bin && \
    rm -rf /flapi ~/go

ENTRYPOINT ["/usr/bin/flapi"]  
