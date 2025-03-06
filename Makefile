<<<<<<< HEAD
.PHONY: build test clean package

# Build the binary
build:
	go build -o talis-agent

# Run tests
test:
	go test ./...

# Clean build artifacts
clean:
	rm -f talis-agent
	rm -rf dist/

# Create deb package
package: build
	mkdir -p dist/DEBIAN
	mkdir -p dist/etc/talis-agent
	mkdir -p dist/usr/local/bin

	# Copy binary
	cp talis-agent dist/usr/local/bin/

	# Copy config file
	cp /etc/talis-agent/config.yaml dist/etc/talis-agent/

	# Create control file
	cat > dist/DEBIAN/control << 'EOL'
Package: talis-agent
Version: 1.0.0
Section: utils
Priority: optional
Architecture: amd64
Maintainer: Celestia <info@celestia.org>
Description: Talis Agent - System Metrics Collection Service
 A Go-based service that monitors system metrics and exposes them via HTTP endpoints.
EOL

	# Create postinst script
	cat > dist/DEBIAN/postinst << 'EOL'
#!/bin/bash
set -e

# Create required directories
mkdir -p /etc/talis-agent/payload

# Set permissions
chmod 755 /usr/local/bin/talis-agent
chmod 755 /etc/talis-agent
chmod 755 /etc/talis-agent/payload
EOL

	chmod 755 dist/DEBIAN/postinst

	# Build the package
	dpkg-deb --build dist talis-agent_1.0.0_amd64.deb 
=======
# Variables
GO_FILES := $(shell find . -name "*.go" -type f)
NIX_FILES := $(shell find . -name "*.nix" -type f)
PROJECTNAME=$(shell basename "$(PWD)")

# Go commands
GO := go
GOTEST := $(GO) test
GOVET := $(GO) vet
GOFMT := gofmt
GOMOD := $(GO) mod
GOBUILD := $(GO) build

# Build flags
LDFLAGS := -ldflags="-s -w"

## help: Get more info on make commands.
help: Makefile
	@echo " Choose a command run in "$(PROJECTNAME)":"
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
.PHONY: help

## all: Run check-env, lint, test, and build
all: lint test build
.PHONY: all

## build: Build the application
build: 
	@echo "Building $(PROJECTNAME)..."
	$(GOBUILD) $(LDFLAGS) -o bin/$(PROJECTNAME) ./cmd/agent/main.go
.PHONY: build

## clean: Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -rf dist/
.PHONY: clean

## test: Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...
.PHONY: test

## fmt: Format code
fmt:
	@echo "Formatting go fmt..."
	$(GOFMT) -w $(GO_FILES)
	@echo "--> Formatting golangci-lint"
	@golangci-lint run --fix
.PHONY: fmt

## lint: Run all linters
lint: fmt vet
	@echo "Running linters..."
	@echo "--> Running golangci-lint"
	@golangci-lint run
	@echo "--> Running actionlint"
	@actionlint
	@echo "--> Running yamllint"
	@yamllint --no-warnings .
.PHONY: lint

## vet: Run go vet
vet:
	@echo "Running go vet..."
	$(GOVET) ./...
.PHONY: vet

## tidy: Tidy and verify dependencies
tidy:
	@echo "Tidying dependencies..."
	$(GOMOD) tidy
	$(GOMOD) verify
.PHONY: tidy

## run: Run the application
run:
	@echo "Running $(PROJECTNAME)..."
	@go run ./cmd/main.go
.PHONY: run

## install-hooks: Install git hooks
install-hooks:
	@echo "Installing git hooks..."
	@git config core.hooksPath .githooks
.PHONY: install-hooks
>>>>>>> f4de01d (feat(ci): adds baseline project CI (#6))
