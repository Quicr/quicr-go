# qgo Makefile - Build orchestration for libquicr Go bindings

.PHONY: all build test bench clean libquicr shim examples lint vet fmt imports check help

# Configuration
LIBQUICR_DIR := libquicr
CSHIM_DIR := cshim
BUILD_DIR := build
INSTALL_DIR := $(BUILD_DIR)/install

# Detect OS
UNAME_S := $(shell uname -s)
NPROC := $(shell nproc 2>/dev/null || sysctl -n hw.ncpu 2>/dev/null || echo 4)

# CGO environment
CGO_CFLAGS := -I$(CURDIR)/$(CSHIM_DIR) -I$(CURDIR)/$(LIBQUICR_DIR)/include
CGO_LDFLAGS := -L$(CURDIR)/$(BUILD_DIR)/lib -lquicr_shim -lquicr -lstdc++ -lm -lpthread

# OpenSSL location
OPENSSL_PREFIX := $(shell brew --prefix openssl@3 2>/dev/null || echo "/opt/homebrew/opt/openssl@3")
# fmt location (required by spdlog)
FMT_PREFIX := $(shell brew --prefix fmt 2>/dev/null || echo "/opt/homebrew/opt/fmt")

# Platform specific flags
ifeq ($(UNAME_S),Darwin)
    CGO_LDFLAGS += -framework Security -framework CoreFoundation
    # Link against dependencies - order matters for static linking
    CGO_LDFLAGS += -L$(CURDIR)/$(BUILD_DIR)/libquicr/dependencies/picoquic -lpicoquic-core -lpicoquic-log -lpicohttp-core
    CGO_LDFLAGS += -L$(CURDIR)/$(BUILD_DIR)/_deps/picotls-build -lpicotls-openssl -lpicotls-minicrypto -lpicotls-core
    CGO_LDFLAGS += -L$(CURDIR)/$(BUILD_DIR)/libquicr/dependencies/spdlog -lspdlog
    CGO_LDFLAGS += -L$(FMT_PREFIX)/lib -lfmt
    CGO_LDFLAGS += -L$(OPENSSL_PREFIX)/lib -lssl -lcrypto
endif

ifeq ($(UNAME_S),Linux)
    CGO_LDFLAGS += -L$(CURDIR)/$(BUILD_DIR)/libquicr/dependencies/picoquic -lpicoquic-core -lpicoquic-log -lpicohttp-core
    CGO_LDFLAGS += -L$(CURDIR)/$(BUILD_DIR)/_deps/picotls-build -lpicotls-openssl -lpicotls-minicrypto -lpicotls-core
    CGO_LDFLAGS += -L$(CURDIR)/$(BUILD_DIR)/libquicr/dependencies/spdlog -lspdlog
    CGO_LDFLAGS += -lssl -lcrypto
endif

# Export CGO environment
export CGO_ENABLED=1
export CGO_CFLAGS
export CGO_LDFLAGS

# Default target
all: shim build

# Help
help:
	@echo "qgo - Go bindings for libquicr"
	@echo ""
	@echo "Targets:"
	@echo "  all       - Build shim and Go package (default)"
	@echo "  shim      - Build libquicr and C shim"
	@echo "  build     - Build Go package"
	@echo "  test      - Run tests"
	@echo "  bench     - Run benchmarks"
	@echo "  examples  - Build example applications"
	@echo "  clean     - Clean build artifacts"
	@echo ""
	@echo "Go tools:"
	@echo "  fmt       - Format Go and C code"
	@echo "  vet       - Run go vet"
	@echo "  lint      - Run linters (golangci-lint or go vet)"
	@echo "  imports   - Fix imports with goimports"
	@echo "  check     - Run fmt, vet, and lint"
	@echo ""

# Initialize submodule if needed
$(LIBQUICR_DIR)/CMakeLists.txt:
	git submodule update --init --recursive

# Build libquicr and C shim
shim: $(LIBQUICR_DIR)/CMakeLists.txt
	@echo "=== Building libquicr and C shim ==="
	@mkdir -p $(BUILD_DIR)
	cd $(BUILD_DIR) && cmake \
		-DCMAKE_BUILD_TYPE=Release \
		-DBUILD_TESTING=OFF \
		-DQUICR_BUILD_TESTS=OFF \
		../$(CSHIM_DIR)
	cd $(BUILD_DIR) && cmake --build . -j$(NPROC)
	@mkdir -p $(BUILD_DIR)/lib
	@cp $(BUILD_DIR)/libquicr_shim.a $(BUILD_DIR)/lib/ 2>/dev/null || true
	@cp $(BUILD_DIR)/libquicr/src/libquicr.a $(BUILD_DIR)/lib/ 2>/dev/null || true
	@echo "=== Build complete ==="

# Check if shim is built
check-shim:
	@if [ ! -f "$(BUILD_DIR)/libquicr_shim.a" ] && [ ! -f "$(BUILD_DIR)/lib/libquicr_shim.a" ]; then \
		echo "Error: C shim not built. Run 'make shim' first."; \
		exit 1; \
	fi

# Build Go package
build: check-shim
	@echo "=== Building Go package ==="
	go build ./...

# Run tests
test: check-shim
	@echo "=== Running tests ==="
	go test -v -race ./...

# Run benchmarks
bench: check-shim
	@echo "=== Running benchmarks ==="
	go test -bench=. -benchmem -run=^$$ ./...

# Build examples
# Note: -ldflags="-s -w" strips debug symbols to fix code signature issues on macOS
examples: check-shim
	@echo "=== Building examples ==="
	@mkdir -p bin
	go build -ldflags="-s -w" -o bin/clock ./examples/clock
	go build -ldflags="-s -w" -o bin/chat ./examples/chat

# Run go vet
vet:
	@echo "=== Running go vet ==="
	go vet ./...

# Fix imports
imports:
	@echo "=== Fixing imports ==="
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w ./qgo ./examples ./internal; \
	else \
		echo "goimports not installed, install with: go install golang.org/x/tools/cmd/goimports@latest"; \
	fi

# Run all checks (fmt, vet, lint)
check: fmt vet
	@echo "=== All checks passed ==="

# Clean build artifacts
clean:
	@echo "=== Cleaning ==="
	rm -rf $(BUILD_DIR)
	rm -rf bin
	go clean ./...


fmt:
	go fmt ./...
	@if command -v clang-format >/dev/null 2>&1; then \
		clang-format -i $(CSHIM_DIR)/*.cpp $(CSHIM_DIR)/*.h; \
	fi
