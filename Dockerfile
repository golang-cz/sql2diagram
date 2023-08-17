# stage 1 building the code
FROM golang:1.21 as builder

COPY / /sql2diagram
WORKDIR /sql2diagram

RUN go mod vendor
RUN go get -u github.com/goware/modvendor
RUN go install github.com/goware/modvendor@latest
RUN modvendor -copy="**/*.c **/*.h **/*.proto" -v
RUN go build -trimpath -ldflags "-s -w" -o sql2diagram ./cmd/sql2diagram/main.go

# stage 2
FROM golang:1.21
# related to https://github.com/golangci/golangci-lint/issues/3107
ENV GOROOT /usr/local/go
# don't place it into $GOPATH/bin because Drone mounts $GOPATH as volume
COPY --from=builder /sql2diagram/sql2diagram /usr/bin/
CMD ["sql2diagram"]
