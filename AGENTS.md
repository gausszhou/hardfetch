# AGENTS.md

This document provides guidelines for AI agents working in this repository.

**语言说明**: 在与用户对话时使用中文。所有代码、命令和文件路径保持英文。

## Project Overview

This is a Go CLI tool called `hardfetch` - a fastfetch/neofetch-like system information tool. The project uses Go 1.22+ and follows standard Go project structure.

## Project Goals

Hardfetch aims to be a fast, customizable system information tool similar to fastfetch/neofetch, but written in Go for better cross-platform compatibility and performance. Key features include:

1. **System Information**: Display OS, kernel, hostname, uptime
2. **Hardware Information**: 
   - CPU: model, cores, threads, frequency, architecture
   - GPU: name, vendor, VRAM, driver version (Windows WMI support)
   - Memory: total, used, available, free
   - Disk: multi-disk support with drive letters (Windows) or mount points
3. **Software Information**: Package manager info, installed packages, services, processes
4. **Network Information**: Local/Public IP, network interfaces
5. **User Information**: Current user, shell, environment variables
6. **Customizable Display**: ASCII logos, color themes, output formatting
7. **Cross-platform**: Support Windows, Linux, macOS
8. **High Performance**: Concurrent information collection
9. **Configurable**: JSON/YAML config files, command-line options
10. **Modular Design**: Select specific modules to display

## Build Commands

### Basic Build
```bash
# Build the binary
make build
# or
go build -o hardfetch cmd/hardfetch/main.go

# Install globally
make install
# or
go install ./cmd/hardfetch

# Clean build artifacts
make clean
```

### Development Build
```bash
# Build with race detector
go build -race -o hardfetch cmd/hardfetch/main.go

# Build for multiple platforms
make build-all

# Build for specific OS/architecture
GOOS=linux GOARCH=amd64 go build -o hardfetch-linux-amd64 cmd/hardfetch/main.go
GOOS=darwin GOARCH=arm64 go build -o hardfetch-darwin-arm64 cmd/hardfetch/main.go
GOOS=windows GOARCH=amd64 go build -o hardfetch-windows-amd64.exe cmd/hardfetch/main.go
```

## Testing Commands

### Run All Tests
```bash
# Run all tests
make test
# or
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests with race detector
go test -race ./...
```

### Run Specific Tests
```bash
# Run tests in a specific package
go test ./internal/cli

# Run a specific test
go test -run TestFunctionName ./internal/cli

# Run tests with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Test Benchmarks
```bash
# Run benchmarks
go test -bench=. ./...

# Run benchmarks with memory profiling
go test -bench=. -benchmem ./...
```

## Linting and Code Quality

### Go Tools
```bash
# Format code
go fmt ./...

# Vet code for suspicious constructs
go vet ./...

# Run static analysis
go install golang.org/x/tools/go/analysis/passes/shadow/cmd/shadow@latest
shadow ./...

# Check for unused dependencies
go mod tidy -v
```

### Recommended Linters
```bash
# Install golangci-lint (if not installed)
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run golangci-lint
golangci-lint run ./...
```

## Code Style Guidelines

### Import Organization
- Use standard library imports first, then third-party imports, then local imports
- Group imports with a blank line between groups
- Use `goimports` to automatically format imports

Example:
```go
import (
    "fmt"
    "os"
    "strings"

    "github.com/spf13/cobra"
    "golang.org/x/text/cases"

    "fe-cli/internal/cli"
)
```

### Naming Conventions
- **Packages**: Use short, lowercase, single-word names (e.g., `cli`, `utils`)
- **Variables**: Use camelCase (e.g., `userName`, `maxRetries`)
- **Constants**: Use CamelCase or UPPER_SNAKE_CASE for exported constants
- **Functions**: Use camelCase; exported functions start with capital letter
- **Interfaces**: Use `-er` suffix when appropriate (e.g., `Reader`, `Writer`)

### Error Handling
- Always check errors immediately after function calls
- Use `fmt.Errorf` with `%w` for wrapping errors
- Provide context in error messages
- Return zero values for errors

Example:
```go
func ReadConfig(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("read config %s: %w", path, err)
    }
    // ... parse config
}
```

### Function Design
- Keep functions small and focused (preferably < 50 lines)
- Return early to reduce nesting
- Use named returns for documentation when helpful
- Document exported functions with complete sentences

### Type Safety
- Use concrete types over `interface{}` when possible
- Define custom types for domain concepts
- Use type assertions with the comma-ok idiom

### Comments and Documentation
- Document all exported functions, types, and variables
- Use complete sentences ending with periods
- Prefer self-documenting code over comments
- Add comments for non-obvious logic

### Project Structure
```
hardfetch/
├── cmd/hardfetch/          # Main application entry point
├── internal/
│   ├── cli/                # CLI-specific logic
│   ├── modules/            # Information collection modules
│   │   ├── system/         # System information
│   │   ├── hardware/       # Hardware information (CPU, GPU, memory, disk)
│   │   ├── software/       # Software information
│   │   ├── network/        # Network information
│   │   └── user/           # User information
│   ├── display/            # Display formatting
│   └── utils/              # Utility functions
├── configs/                # Configuration files
├── logos/                  # ASCII logo files
├── go.mod                  # Go module definition
├── Makefile                # Build automation
└── README.md               # Project documentation
```

## Development Workflow

### Adding New Features
1. Create appropriate directory structure under `internal/` or `pkg/`
2. Write tests for new functionality
3. Update `cmd/fe-cli/main.go` to integrate new features
4. Run tests and ensure they pass
5. Update documentation if needed

### Adding Dependencies
```bash
# Add a new dependency
go get github.com/example/package

# Update all dependencies
go get -u ./...

# Clean up unused dependencies
go mod tidy
```

### Version Management
- Update version in `internal/cli/version.go`
- Use semantic versioning (MAJOR.MINOR.PATCH)
- Tag releases with `git tag v0.1.0`

## Git Guidelines

### Commit Messages
- Use conventional commits format: `type(scope): description`
- Types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`
- Keep first line under 50 characters
- Provide detailed description in commit body when needed

### Branch Strategy
- `main`: Production-ready code
- `develop`: Integration branch for features
- Feature branches: `feature/description`
- Bug fix branches: `fix/description`

## Performance Considerations

### Build Optimization
```bash
# Build with optimizations
go build -ldflags="-s -w" -o fe-cli cmd/fe-cli/main.go

# Strip debug information for smaller binary
go build -trimpath -o fe-cli cmd/fe-cli/main.go
```

### Runtime Performance
- Avoid unnecessary allocations in hot paths
- Use `sync.Pool` for frequently allocated objects
- Profile with `go tool pprof` when performance is critical

## Common Tasks

### Adding a New Module
1. Create module implementation in `internal/modules/`
2. Add module to the appropriate category (system, hardware, software, network, user)
3. Update `cmd/hardfetch/main.go` to include the module in display functions
4. Write tests for the module
5. Update help text and documentation

### Debugging
```bash
# Run with debug logging
DEBUG=1 ./fe-cli

# Use delve debugger
dlv debug cmd/fe-cli/main.go
```

### Cross-Compilation
```bash
# Build for multiple platforms
GOOS=darwin GOARCH=arm64 go build -o fe-cli-darwin-arm64 cmd/fe-cli/main.go
GOOS=linux GOARCH=amd64 go build -o fe-cli-linux-amd64 cmd/fe-cli/main.go
GOOS=windows GOARCH=amd64 go build -o fe-cli-windows-amd64.exe cmd/fe-cli/main.go
```

## Quality Assurance Checklist

Before committing code:
- [ ] All tests pass
- [ ] Code is formatted with `go fmt`
- [ ] No `go vet` warnings
- [ ] No linting issues
- [ ] Documentation updated if needed
- [ ] Backward compatibility maintained

## Resources

- [Go Documentation](https://golang.org/doc/)
- [Effective Go](https://golang.org/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Go Proverbs](https://go-proverbs.github.io/)