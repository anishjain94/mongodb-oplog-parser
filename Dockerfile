FROM golang:alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o oplog_parser .

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/oplog_parser .

CMD ["./oplog_parser"]