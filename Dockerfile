# builder
FROM golang:1.25 as builder

COPY / /sql2diagram
WORKDIR /sql2diagram

RUN go build -trimpath -ldflags "-s -w" -o sql2diagram ./cmd/sql2diagram

# runner
FROM golang:1.25
COPY --from=builder /sql2diagram/sql2diagram /usr/bin/
CMD ["sql2diagram"]
