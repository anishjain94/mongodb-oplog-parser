FROM --platform=$BUILDPLATFORM golang:alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Use ARG for specifying target platform
ARG TARGETPLATFORM
ARG BUILDPLATFORM
RUN echo "I am running on $BUILDPLATFORM, building for $TARGETPLATFORM"

# Build the application
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o oplog_parser .

# Use a multi-arch base image for the final stage
FROM --platform=$TARGETPLATFORM alpine:latest

WORKDIR /root/

COPY --from=builder /app/oplog_parser .

CMD ["./oplog_parser"]
