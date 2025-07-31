.PHONY: all build test clean fmt lint

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOCLEAN=$(GOCMD) clean
GOGET=$(GOCMD) get
GOFMT=gofmt
GOLINT=golangci-lint

# Build variables
BINARY_NAME=go-agent-kit
BINARY_DIR=dist

all: test build

build:
	$(GOBUILD) -v ./...

test:
	$(GOTEST) -v -race -coverprofile=coverage.out ./...

test-short:
	$(GOTEST) -v -short ./...

coverage:
	$(GOTEST) -race -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

clean:
	$(GOCLEAN)
	rm -rf $(BINARY_DIR)
	rm -f coverage.out coverage.html

fmt:
	$(GOFMT) -s -w .

lint:
	$(GOLINT) run

deps:
	$(GOGET) -u ./...
	$(GOCMD) mod tidy

mod-download:
	$(GOCMD) mod download

mod-verify:
	$(GOCMD) mod verify

# Development helpers
dev-test:
	$(GOTEST) -v -count=1 ./...

benchmark:
	$(GOTEST) -bench=. -benchmem ./...