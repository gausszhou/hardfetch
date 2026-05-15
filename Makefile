.PHONY: build test clean install build-all lint fmt vet

BINARY_NAME=hardfetch
DIST_DIR=dist

build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -o $(DIST_DIR)/$(BINARY_NAME) main.go

test:
	go test ./...

clean:
	rm -rf $(DIST_DIR)
	rm -f coverage.out

install:
	@echo "Installing $(BINARY_NAME)..."
	CGO_ENABLED=0 go install -ldflags="-s -w" .
	@echo "$(BINARY_NAME) installed successfully to $$(go env GOPATH)/bin"

build-all: build-linux build-darwin build-windows

build-linux:
	mkdir -p $(DIST_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64 main.go
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o $(DIST_DIR)/$(BINARY_NAME)-linux-arm64 main.go

build-darwin:
	mkdir -p $(DIST_DIR)
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64 main.go
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64 main.go

build-windows:
	mkdir -p $(DIST_DIR)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o $(DIST_DIR)/$(BINARY_NAME)-windows-amd64.exe main.go
	CGO_ENABLED=0 GOOS=windows GOARCH=arm64 go build -ldflags="-s -w" -o $(DIST_DIR)/$(BINARY_NAME)-windows-arm64.exe main.go

lint:
	@which golangci-lint >/dev/null 2>&1 || go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	golangci-lint run ./...

fmt:
	go fmt ./...

vet:
	go vet ./...