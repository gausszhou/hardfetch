# hardfetch

A Go CLI tool for fetching system information, similar to fastfetch/neofetch.

## Features

- System, hardware, network, and battery information
- Cross-platform support (Windows/Linux/macOS)
- High-performance concurrent data collection
- Configurable display

## Installation

```bash
go install github.com/gausszhou/hardfetch@latest
```

Or build from source:

```bash
git clone https://github.com/gausszhou/hardfetch.git
cd hardfetch
make install
```

## Usage

```bash
hardfetch           # Default display
hardfetch -d        # Debug mode
hardfetch -v        # Version
hardfetch -h        # Help
```

## Build

```bash
make build   # Build
make test    # Test
make install # Install
```
