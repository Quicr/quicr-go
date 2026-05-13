#!/bin/bash
set -e

# Download pre-built shim libraries for qgo
# Usage: ./scripts/download-shim.sh [version]

REPO="quicr/qgo"
VERSION="${1:-latest}"

# Detect platform
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$OS" in
    darwin) OS="darwin" ;;
    linux) OS="linux" ;;
    *)
        echo "Error: Unsupported OS: $OS"
        exit 1
        ;;
esac

case "$ARCH" in
    x86_64|amd64) ARCH="amd64" ;;
    arm64|aarch64) ARCH="arm64" ;;
    *)
        echo "Error: Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

ARTIFACT="qgo-libs-${OS}-${ARCH}.tar.gz"
echo "Platform: ${OS}-${ARCH}"

# Get download URL
if [ "$VERSION" = "latest" ]; then
    echo "Fetching latest release..."
    DOWNLOAD_URL=$(curl -s "https://api.github.com/repos/${REPO}/releases/latest" | \
        grep "browser_download_url.*${ARTIFACT}" | \
        cut -d '"' -f 4)
else
    echo "Fetching release ${VERSION}..."
    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${ARTIFACT}"
fi

if [ -z "$DOWNLOAD_URL" ]; then
    echo "Error: Could not find pre-built libraries for ${OS}-${ARCH}"
    echo "You may need to build from source: make shim"
    exit 1
fi

echo "Downloading: ${DOWNLOAD_URL}"

# Create build directories matching the expected layout
mkdir -p build/lib
mkdir -p build/libquicr/dependencies/picoquic
mkdir -p build/libquicr/dependencies/spdlog
mkdir -p build/_deps/picotls-build

# Download and extract to temp dir first
TMPFILE=$(mktemp)
TMPDIR=$(mktemp -d)
curl -L -o "$TMPFILE" "$DOWNLOAD_URL"

echo "Extracting libraries..."
tar -xzvf "$TMPFILE" -C "$TMPDIR"

# Copy libraries to expected locations
cp "$TMPDIR"/libquicr_shim.a "$TMPDIR"/libquicr.a build/lib/
cp "$TMPDIR"/libpicoquic-core.a "$TMPDIR"/libpicoquic-log.a "$TMPDIR"/libpicohttp-core.a build/libquicr/dependencies/picoquic/
cp "$TMPDIR"/libspdlog.a build/libquicr/dependencies/spdlog/
cp "$TMPDIR"/libpicotls-*.a build/_deps/picotls-build/

rm -rf "$TMPFILE" "$TMPDIR"

echo ""
echo "Successfully downloaded pre-built libraries!"
echo "You can now build your project with: go build ./..."
