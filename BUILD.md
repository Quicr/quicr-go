# Building from Source

This document covers building qgo from source code.

## Requirements

- Go 1.21 or later
- C++17 compatible compiler (clang++ or g++)
- CMake 3.15 or later
- Git (for submodules)

### Platform Dependencies

**macOS:**
```bash
brew install cmake openssl@3 fmt
```

**Linux (Ubuntu/Debian):**
```bash
sudo apt-get install cmake build-essential libssl-dev libfmt-dev
```

## Clone Repository

```bash
git clone --recursive https://github.com/quicr/qgo.git
cd qgo
```

If you already cloned without `--recursive`:
```bash
git submodule update --init --recursive
```

## Build Steps

### 1. Build the C Shim and libquicr

```bash
make shim
```

This builds:
- libquicr (the C++ MoQT library)
- quicr_shim (the CGO-compatible C wrapper)

### 2. Build the Go Package

```bash
make build
```

### 3. Run Tests

```bash
make test
```

### 4. Run Benchmarks

```bash
make bench
```

### 5. Build Examples

```bash
make examples
```

## Makefile Targets

| Target | Description |
|--------|-------------|
| `make shim` | Build C shim and libquicr |
| `make build` | Build Go package |
| `make test` | Run tests |
| `make bench` | Run benchmarks |
| `make examples` | Build example applications |
| `make clean` | Clean build artifacts |
| `make fmt` | Format Go code |
| `make vet` | Run go vet |
| `make lint` | Run golangci-lint |

## Upgrading libquicr

To update to the latest libquicr version:

```bash
cd libquicr
git fetch origin
git checkout origin/main
cd ..

make clean
make shim
make build
make test
```

## Troubleshooting

### OpenSSL not found (macOS)

If CMake can't find OpenSSL:
```bash
export OPENSSL_ROOT_DIR=$(brew --prefix openssl@3)
make clean && make shim
```

### Linker errors

Ensure the shim is built before the Go package:
```bash
make clean
make shim
make build
```
