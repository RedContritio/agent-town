#!/bin/bash
# Setup all dependencies

set -e

echo "=== Agent Town Setup ==="
echo ""

# Check if already installed
check_installed() {
    if command -v go &> /dev/null && command -v godot4 &> /dev/null; then
        echo "✓ Dependencies already installed"
        echo "  Go: $(go version | awk '{print $3}')"
        echo "  Godot: $(godot4 --version | head -1)"
        return 0
    fi
    return 1
}

if check_installed; then
    echo ""
    read -p "Reinstall? [y/N] " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Skipping setup"
        exit 0
    fi
fi

# Install Go
echo "→ Installing Go..."
"$(dirname "$0")/install-go.sh"

# Install Godot
echo ""
echo "→ Installing Godot..."
"$(dirname "$0")/install-godot.sh"

# Install templates (optional)
echo ""
read -p "Install Godot export templates (~1GB download)? [y/N] " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    "$(dirname "$0")/install-godot-templates.sh"
else
    echo "Skipped. Install later with: ./scripts/install-godot-templates.sh"
fi

echo ""
echo "✓ Setup complete!"
echo "  Next: make build && make web && make run"
