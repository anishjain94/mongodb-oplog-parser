#TODO: How to ensure the image runs on both ARM and AMD64 architectures
FROM golang:alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o oplog_parser .
RUN go test -v ./transformer 

RUN go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
RUN golangci-lint run

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/oplog_parser .

CMD ["./oplog_parser"]