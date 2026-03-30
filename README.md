# HardFetch

A command-line tool for fetching system information, similar to fastfetch/neofetch.

## Features

- System, hardware, network, and battery information
- Cross-platform support (Windows/Linux/macOS)
- High-performance concurrent data collection

## Installation

```bash
# Install latest version
go install github.com/gausszhou/hardfetch@latest

# Or build from source
git clone https://github.com/gausszhou/hardfetch.git
cd hardfetch
make install
```

## Usage

```bash
# Run the tool
hardfetch

# Check version
hardfetch --version
```

## Clean

```bash
# Clean build artifacts
make clean

# Clean Go module cache (all cached versions)
go clean -cache

# Remove installed binary
rm -f $(go env GOPATH)/bin/hardfetch
```

## License

MIT License
