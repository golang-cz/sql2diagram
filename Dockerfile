# stage 1 building the code
FROM golang:1.25 as builder

COPY / /sql2diagram
WORKDIR /sql2diagram

RUN go build -trimpath -ldflags "-s -w" -o sql2diagram ./main.go

# stage 2
FROM golang:1.25
# related to https://github.com/golangci/golangci-lint/issues/3107
ENV GOROOT /usr/local/go
# don't place it into $GOPATH/bin because Drone mounts $GOPATH as volume
COPY --from=builder /sql2diagram/sql2diagram /usr/bin/
CMD ["sql2diagram"]
