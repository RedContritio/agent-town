#!/bin/bash
# Install Godot 4.5+ via snap

set -e

check_godot() {
    for cmd in godot4 godot /snap/bin/godot4; do
        if command -v $cmd &> /dev/null; then
            VERSION=$($cmd --version 2>/dev/null | head -1)
            echo "✓ Godot already installed: $VERSION"
            return 0
        fi
    done
    return 1
}

install_godot() {
    echo "→ Installing Godot..."
    
    if ! command -v snap &> /dev/null; then
        echo "✗ snap not found. Install snapd:"
        echo "  sudo apt-get install snapd"
        exit 1
    fi
    
    echo "  Installing via snap..."
    sudo snap install godot4
    
    echo "✓ Godot installed"
}

main() {
    if check_godot; then exit 0; fi
    install_godot
    
    GODOT_PATH=$(command -v godot4 || echo /snap/bin/godot4)
    if [ -x "$GODOT_PATH" ]; then
        echo "✓ Verification: $($GODOT_PATH --version | head -1)"
    else
        echo "✗ Installation failed"
        exit 1
    fi
}

main "$@"
