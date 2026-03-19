#!/bin/bash
# Install Go 1.22+ if not present

set -e

GO_VERSION="1.22.0"
GO_INSTALL_DIR="/usr/local"

check_go() {
    if command -v go &> /dev/null; then
        INSTALLED_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
        echo "✓ Go already installed: $INSTALLED_VERSION"
        return 0
    fi
    return 1
}

install_go() {
    echo "→ Installing Go $GO_VERSION..."
    
    ARCH=$(uname -m)
    case $ARCH in
        x86_64) GO_ARCH="amd64" ;;
        aarch64|arm64) GO_ARCH="arm64" ;;
        *) echo "✗ Unsupported architecture: $ARCH"; exit 1 ;;
    esac
    
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    DOWNLOAD_URL="https://go.dev/dl/go${GO_VERSION}.${OS}-${GO_ARCH}.tar.gz"
    TEMP_DIR=$(mktemp -d)
    
    echo "  Downloading..."
    curl -fsSL "$DOWNLOAD_URL" -o "$TEMP_DIR/go.tar.gz"
    
    echo "  Extracting..."
    sudo rm -rf "$GO_INSTALL_DIR/go"
    sudo tar -C "$GO_INSTALL_DIR" -xzf "$TEMP_DIR/go.tar.gz"
    rm -rf "$TEMP_DIR"
    
    if ! grep -q "/usr/local/go/bin" ~/.bashrc 2>/dev/null; then
        echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
    fi
    export PATH=$PATH:/usr/local/go/bin
    
    echo "✓ Go $GO_VERSION installed"
}

main() {
    if check_go; then exit 0; fi
    install_go
    
    if command -v go &> /dev/null; then
        echo "✓ Verification: $(go version)"
    else
        echo "✗ Installation failed"
        exit 1
    fi
}

main "$@"
