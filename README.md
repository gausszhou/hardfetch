# hardfetch

A fastfetch/neofetch-like system information tool written in Go.

## Features

- System information display (OS, kernel, hostname, uptime)
- Hardware information (CPU, memory, disk)
- Software information (package managers, processes)
- Network information (IP addresses, interfaces)
- Customizable ASCII logos and color themes
- Cross-platform support (Windows, Linux, macOS)
- High performance with concurrent information collection
- Configurable via JSON/YAML files

## Installation

### From Source

```bash
# Clone the repository
git clone <repository-url>
cd hardfetch

# Build and install
make install
# or
go install ./cmd/hardfetch
```

### Using go install

```bash
go install hardfetch/cmd/hardfetch@latest
```

## Usage

```bash
# Show system information with default settings
hardfetch

# Show specific modules
hardfetch --modules system,cpu,memory

# Show all available modules
hardfetch --all

# Show without ASCII logo
hardfetch --no-logo

# Show without colors
hardfetch --no-colors

# Show version
hardfetch --version

# Show help
hardfetch --help

# Generate config file
hardfetch --gen-config

# List all available modules
hardfetch --list-modules
```

## Available Modules

- **system**: Operating system, kernel, hostname, uptime, CPU cores
- **cpu**: CPU model, architecture, cores, threads, frequency
- **memory**: Total, used, available, and free memory
- **disk**: Disk usage information
- **network**: Hostname, local IP, network interfaces
- **software**: Shell, editor, Go version, package managers, process count

## Development

### Build Commands

```bash
# Build binary to dist/ directory
make build

# Run tests
make test

# Clean build artifacts (removes dist/ directory)
make clean

# Install globally
make install

# Build for multiple platforms (Linux, macOS, Windows)
make build-all

# Build for specific platform
make build-linux    # Build Linux binaries
make build-darwin   # Build macOS binaries  
make build-windows  # Build Windows binaries
```

### Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific test
go test -run TestMainVersion ./cmd/fe-cli
```

### Code Quality

```bash
# Format code
go fmt ./...

# Vet code
go vet ./...

# Run linter (requires golangci-lint)
golangci-lint run ./...
```

## Project Structure

```
hardfetch/
├── cmd/hardfetch/          # Main application entry point
│   ├── main.go             # CLI entry point
│   └── main_test.go        # CLI tests
├── internal/
│   ├── cli/                # CLI-specific logic
│   │   ├── version.go      # Version constant
│   │   ├── config.go       # Configuration management
│   │   └── ...
│   ├── modules/            # Information collection modules
│   │   ├── system/         # System information
│   │   ├── hardware/       # Hardware information
│   │   ├── software/       # Software information
│   │   ├── network/        # Network information
│   │   └── user/           # User information
│   ├── display/            # Display formatting
│   │   ├── ascii.go        # ASCII art rendering
│   │   ├── colors.go       # Color support
│   │   └── formatter.go    # Output formatting
│   └── utils/              # Utility functions
├── dist/                   # Build outputs (generated, gitignored)
├── configs/                # Configuration files
│   └── default.json        # Default configuration
├── logos/                  # ASCII logo files
├── go.mod                  # Go module definition
├── go.sum                  # Dependency checksums
├── Makefile                # Build automation
├── .gitignore              # Git ignore rules
├── .golangci.yml           # Linter configuration
└── README.md               # Project documentation
```

## Adding New Features

1. Create appropriate directory structure under `internal/`
2. Write tests for new functionality
3. Update `cmd/fe-cli/main.go` to integrate new features
4. Run tests and ensure they pass
5. Update documentation if needed

## Code Style

Follow the guidelines in [AGENTS.md](AGENTS.md) for:
- Import organization
- Naming conventions
- Error handling
- Function design
- Type safety
- Comments and documentation

## License

MIT