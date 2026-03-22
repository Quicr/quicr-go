<p align="center">
  <img src="logo.svg" alt="qgo logo" width="150">
</p>

# qgo - Go Bindings for libquicr

[![CI](https://github.com/Quicr/moq-go/actions/workflows/ci.yml/badge.svg)](https://github.com/Quicr/moq-go/actions/workflows/ci.yml)

`qgo` provides idiomatic Go bindings for [libquicr](https://github.com/Quicr/libquicr), a C++ implementation of the Media over QUIC Transport (MoQT) protocol.

**Supports MoQT draft-16 via libquicr.**


## Requirements

- Go 1.21 or later
- C++17 compatible compiler (clang++ or g++)
- CMake 3.15 or later
- Git (for submodules)

### Platform Dependencies

**macOS:**
```bash
# Install via Homebrew
brew install cmake openssl@3 fmt
```

**Linux (Ubuntu/Debian):**
```bash
sudo apt-get install cmake build-essential libssl-dev libfmt-dev
```

## Installation

### 1. Clone with Submodules

```bash
git clone --recursive https://github.com/quicr/qgo.git
cd qgo
```

If you already cloned without `--recursive`:
```bash
git submodule update --init --recursive
```

### 2. Build the C Shim and libquicr

```bash
make shim
```

This builds:
- libquicr (the C++ library)
- quicr_shim (the CGO-compatible C wrapper)

### 3. Build the Go Package

```bash
make build
```

### 4. Run Tests

```bash
make test
```

### 5. Build Examples

```bash
make examples
```

## Upgrading libquicr

To update to the latest libquicr version:

```bash
# Update submodule to latest
cd libquicr
git fetch origin
git checkout origin/main
cd ..

# Clean and rebuild
make clean
make shim
make build
make test
```

## Examples

### Clock Example

Publishes or subscribes to UTC timestamps.

```bash
# Build
make examples

# Publish timestamps
./bin/clock -mode publish -server localhost:4433

# Subscribe to timestamps (in another terminal)
./bin/clock -mode subscribe -server localhost:4433
```

### Chat Example

Multi-user chat using namespace subscriptions.

```bash
# Build
make examples

# Start chat (requires a MoQ relay server)
./bin/chat -server localhost:4433 -session my-room
```

See [examples/chat/README.md](examples/chat/README.md) for details.

## License

This project is licensed under the same terms as libquicr. See [LICENSE](LICENSE) for details.
