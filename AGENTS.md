# AGENTS.md - Agentic Coding Guidelines

## Project Overview

This is a **Go** project (go 1.24.0) that fetches hardware/system info using [gopsutil](https://github.com/shirou/gopsutil).

## Build / Lint / Test Commands

### Build

```bash
make build          # Build for current platform
make build-all     # Build for all platforms (linux, darwin, windows)
make build-linux   # Build for Linux (amd64, arm64)
make build-darwin  # Build for macOS (amd64, arm64)
make build-windows # Build for Windows (amd64, arm64)
make install       # Install binary to $GOPATH/bin
make clean         # Clean build artifacts
```

### Testing

```bash
make test                    # Run all tests
go test ./...               # Run all tests
go test -run TestName ./... # Run a single test
```

### Linting & Formatting

```bash
make lint  # Run golangci-lint
make fmt   # Format code with go fmt
make vet   # Run go vet
```

---

## Code Style Guidelines

### General

- Follow standard Go conventions (see [Effective Go](https://go.dev/doc/effective_go))
- Keep functions small and focused

### Imports

- Use grouped imports with `goimports`
- Local packages: `github.com/gausszhou/hardfetch`
- Third-party: `github.com/shirou/gopsutil/v4/cpu`

```go
import (
    "fmt"

    "github.com/shirou/gopsutil/v4/cpu"

    "github.com/gausszhou/hardfetch/internal/logger"
)
```

### Naming Conventions

- **Files**: snake_case (e.g., `cpu_info.go`, `detect_platform.go`)
- **Types/Exported Functions**: PascalCase (e.g., `Get()`, `Detect()`)
- **Variables/Unexported Functions**: camelCase (e.g., `resultOnce`)
- **Interfaces**: Name with `er` suffix (e.g., `Detector`)

### Types & Declarations

- Use explicit type declarations for exported types
- Use pointers (`*Type`) for mutable objects

```go
type Info struct {
    Model        string
    Cores        int
    Threads      int
    Frequency    string
}
```

### Error Handling

- Always check errors with `if err != nil`
- Return errors from functions (don't panic)
- Use context for cancellation/timeouts

```go
func Get() (*Info, error) {
    info := &Info{}
    cpuInfo, err := cpu.Info()
    if err != nil {
        return nil, err
    }
    return info, nil
}
```

### Logging

- Use `internal/logger` package
- Use `logger.Debug()` for debugging
- Use `logger.StartTimer()` for performance measurement

### Concurrency

- Use `sync.WaitGroup` for synchronization
- Use `sync.Once` for one-time initialization
- Use `context.Context` for cancellation

### Project Structure

```
hardfetch/
├── main.go              # Entry point
├── main_test.go         # Tests
├── internal/
│   ├── cli/             # CLI constants
│   ├── detect/          # Detection logic
│   ├── display/         # Output display
│   ├── logger/          # Logging
│   └── modules/         # Hardware info modules
│       ├── battery/
│       ├── cpuinfo/
│       ├── disk/
│       ├── gpuinfo/
│       ├── memory/
│       ├── network/
│       └── sys/
```

### Testing

- Write tests in `*_test.go` files
- Use table-driven tests when appropriate
- Test both success and error cases

### Linter Configuration

The project uses golangci-lint with these linters:
- errcheck, gosimple, govet, ineffassign, staticcheck, typecheck, unused
- gofmt (simplify: true), goimports, revive

Linter rules:
- All exported functions must have comments (revive `exported`)
- Package-level comments required (revive `package-comments`)

---

## Common Tasks

### Adding a New Module

1. Create `internal/modules/<modulename>/<modulename>.go`
2. Implement `Get() (*Info, error)` function
3. Register detector in `internal/detect/detect.go`

### Adding a CLI Flag

Edit `main.go` - add your flag in the switch statement:

```go
case "--flag", "-f":
    // handle flag
```

### Debug Mode

```bash
./hardfetch -d   # or --debug
```
