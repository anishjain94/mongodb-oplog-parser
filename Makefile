# Determine the operating system
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)

# Binary name
BINARY_NAME := grep
ifeq ($(GOOS),windows)
    BINARY_NAME := $(BINARY_NAME).exe
endif

.PHONY: all build clean test

all: build

build:
	@echo "Building binary for $(GOOS)/$(GOARCH)..."
	go build -o $(BINARY_NAME) .

clean:
	@echo "Cleaning up..."
	go clean
	rm -f $(BINARY_NAME)

test:
	@echo "Running tests..."
	go test -v ./...