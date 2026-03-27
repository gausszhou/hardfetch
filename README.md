# hardfetch

A fastfetch/neofetch-like system information tool written in Go.

## Features

- **System Information**: OS, kernel, hostname, uptime
- **Hardware Information**: 
  - CPU: model, cores, threads, frequency, architecture
  - GPU: name, vendor, VRAM, driver version (Windows WMI support)
  - Memory: total, used, available, free
  - Disk: multi-disk support with drive letters (Windows) or mount points
- **Network Information**: hostname, local IP, public IP, network interfaces
- **Customizable Display**: ASCII logos, color themes, output formatting
- **Cross-platform**: Support Windows, Linux, macOS
- **High Performance**: Concurrent information collection
- **Configurable**: JSON/YAML config files, command-line options
- **Modular Design**: Select specific modules to display

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

- **system**: Operating system, kernel, hostname, uptime
- **cpu**: CPU model, architecture, cores, threads, frequency
- **gpu**: GPU name, vendor, VRAM, driver version
- **memory**: Total, used, available, and free memory
- **disk**: Multi-disk information with drive letters/mount points
- **network**: Hostname, local IP, public IP, network interfaces

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
│   │   ├── hardware/       # Hardware information (CPU, GPU, memory, disk)
│   │   │   ├── hardware.go          # Common hardware interface
│   │   │   ├── hardware_windows.go  # Windows-specific implementations
│   │   │   └── hardware_other.go    # Non-Windows implementations
│   │   └── network/        # Network information
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
├── AGENTS.md               # Development guidelines
└── README.md               # Project documentation
```

## Platform-Specific Implementations

### Windows
- **CPU Detection**: Registry queries for CPU model and frequency
- **GPU Detection**: WMI queries (Win32_VideoController) for GPU information
- **Disk Detection**: Multi-disk support with drive letters using Windows API
- **Memory**: Windows GlobalMemoryStatusEx API

### Linux/macOS
- **CPU Detection**: Generic implementation with runtime.NumCPU()
- **GPU Detection**: OS-specific placeholder implementations
- **Disk Detection**: Single disk placeholder
- **Memory**: Placeholder values (to be implemented with system-specific calls)

## GPU Information Detection

The GPU module provides detailed graphics card information:

### Windows Implementation
- Uses Windows Management Instrumentation (WMI) via PowerShell
- Queries `Win32_VideoController` class for GPU details
- Extracts: Name, Vendor, VRAM, Driver Version
- Supports multiple GPUs
- Automatic vendor detection (NVIDIA, AMD, Intel, Microsoft)

### Example Output
```
GPU Information:
----------------
GPU:
  Name           : NVIDIA GeForce RTX 4060 Laptop GPU
  Vendor         : NVIDIA
  VRAM           : 4.00 GiB
  Driver         : 32.0.15.7283
```

## Example Output

```bash
$ hardfetch --all

System Information:
------------------
OS             : windows
Arch           : amd64
Kernel         : Windows
Hostname       : DESKTOP-EXAMPLE
Uptime         : 2 days, 5 hours

CPU Information:
----------------
Model          : Intel Core i7-12700K
Cores          : 12
Threads        : 20
Frequency      : 3.60 GHz
Architecture   : x64

GPU Information:
----------------
GPU:
  Name           : NVIDIA GeForce RTX 3070
  Vendor         : NVIDIA
  VRAM           : 8.00 GiB
  Driver         : 31.0.15.5123

Memory Information:
-------------------
Total          : 32.00 GiB
Used           : 8.42 GiB
Available      : 23.58 GiB
Free           : 23.58 GiB

Disk Information:
-----------------
Drive C::
  Total          : 512.00 GiB
  Used           : 256.42 GiB
  Free           : 255.58 GiB

Drive D::
  Total          : 1.00 TiB
  Used           : 512.34 GiB
  Free           : 511.66 GiB

Network Information:
--------------------
Hostname       : DESKTOP-EXAMPLE
Local IP       : 192.168.1.100
Public IP      : 203.0.113.1
Interfaces     : Ethernet (192.168.1.100), Wi-Fi (192.168.1.101)
```

## Adding New Features

1. Create appropriate directory structure under `internal/`
2. Write tests for new functionality
3. Update `cmd/hardfetch/main.go` to integrate new features
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