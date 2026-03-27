.PHONY: build test clean install build-all lint fmt vet

BINARY_NAME=hardfetch
DIST_DIR=dist

build:
	go build -o $(DIST_DIR)/$(BINARY_NAME) cmd/hardfetch/main.go

test:
	go test ./...

clean:
	rm -rf $(DIST_DIR)
	rm -f coverage.out

install:
	go install ./cmd/hardfetch

build-all: build-linux build-darwin build-windows

build-linux:
	mkdir -p $(DIST_DIR)
	GOOS=linux GOARCH=amd64 go build -o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64 cmd/hardfetch/main.go
	GOOS=linux GOARCH=arm64 go build -o $(DIST_DIR)/$(BINARY_NAME)-linux-arm64 cmd/hardfetch/main.go

build-darwin:
	mkdir -p $(DIST_DIR)
	GOOS=darwin GOARCH=amd64 go build -o $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64 cmd/hardfetch/main.go
	GOOS=darwin GOARCH=arm64 go build -o $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64 cmd/hardfetch/main.go

build-windows:
	mkdir -p $(DIST_DIR)
	GOOS=windows GOARCH=amd64 go build -o $(DIST_DIR)/$(BINARY_NAME)-windows-amd64.exe cmd/hardfetch/main.go
	GOOS=windows GOARCH=arm64 go build -o $(DIST_DIR)/$(BINARY_NAME)-windows-arm64.exe cmd/hardfetch/main.go

lint:
	golangci-lint run ./...

fmt:
	go fmt ./...

vet:
	go vet ./...