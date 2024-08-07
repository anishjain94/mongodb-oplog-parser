# TODO: add multi-stage build
# TODO: use an alpine golang image
# TODO: ensure that docker image runs in arm and well as amd64.

FROM golang:latest 

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o oplog_parser .
RUN go test -v ./ ./transformer 

RUN go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
RUN golangci-lint run

CMD ["./oplog_parser"]