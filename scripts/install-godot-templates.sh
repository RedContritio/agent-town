#!/bin/bash
# Install Godot 4.5 export templates

set -e

GODOT_VERSION="4.5"
TEMPLATE_DIR="$HOME/.local/share/godot/export_templates/${GODOT_VERSION}.stable"
MIRROR=""

check_templates() {
    if [ -f "$TEMPLATE_DIR/web_release.zip" ]; then
        echo "✓ Export templates already installed"
        return 0
    fi
    return 1
}

find_godot() {
    for cmd in godot4 godot /snap/bin/godot4; do
        if command -v $cmd &> /dev/null; then
            echo $cmd
            return
        fi
    done
    echo ""
}

# Try to find fastest mirror
detect_mirror() {
    echo "→ Testing download mirrors..."
    
    MIRRORS=(
        "https://github.com/godotengine/godot/releases/download"  # GitHub 在国内某些地区更快
        "https://downloads.tuxfamily.org/godotengine"              # 官方源
    )
    
    BEST_MIRROR=""
    BEST_TIME=999
    
    for mirror in "${MIRRORS[@]}"; do
        TIME=$(curl -o /dev/null -s -w "%{time_total}" --max-time 5 -I "$mirror" 2>/dev/null || echo "999")
        if (( $(echo "$TIME < $BEST_TIME" | bc -l 2>/dev/null || echo "0") )); then
            BEST_TIME=$TIME
            BEST_MIRROR=$mirror
        fi
    done
    
    if [ -n "$BEST_MIRROR" ] && [ "$BEST_TIME" != "999" ]; then
        echo "  Using: $BEST_MIRROR (${BEST_TIME}s)"
        MIRROR="$BEST_MIRROR"
    else
        echo "  Using default mirror"
        MIRROR="https://downloads.tuxfamily.org/godotengine"
    fi
}

download_templates() {
    local url=$1
    local output=$2
    local max_retries=3
    local retry=0
    
    while [ $retry -lt $max_retries ]; do
        echo "  Downloading (attempt $((retry+1))/$max_retries)..."
        
        if curl -fsSL --continue-at - --progress-bar -o "$output.tmp" "$url"; then
            mv "$output.tmp" "$output"
            echo "  ✓ Download complete"
            return 0
        fi
        
        retry=$((retry+1))
        if [ $retry -lt $max_retries ]; then
            echo "  Failed, retrying in 3s..."
            sleep 3
        fi
    done
    
    rm -f "$output.tmp"
    return 1
}

install_templates_manual() {
    echo "→ Installing Godot $GODOT_VERSION export templates..."
    
    detect_mirror
    mkdir -p "$TEMPLATE_DIR"
    
    # Construct download URL based on mirror type
    if echo "$MIRROR" | grep -q "github.com"; then
        # GitHub URL format: https://github.com/godotengine/godot/releases/download/4.5-stable/Godot_v4.5-stable_export_templates.tpz
        VERSION_URL="$MIRROR/${GODOT_VERSION}-stable"
    else
        # Tuxfamily URL format: https://downloads.tuxfamily.org/godotengine/4.5/
        VERSION_URL="$MIRROR/$GODOT_VERSION"
    fi
    TEMPLATE_FILE="Godot_v${GODOT_VERSION}-stable_export_templates.tpz"
    DOWNLOAD_URL="$VERSION_URL/$TEMPLATE_FILE"
    TEMP_FILE="/tmp/$TEMPLATE_FILE"
    
    echo "  URL: $DOWNLOAD_URL"
    
    if ! download_templates "$DOWNLOAD_URL" "$TEMP_FILE"; then
        echo "✗ Download failed"
        return 1
    fi
    
    echo "  Extracting..."
    unzip -q "$TEMP_FILE" -d "$TEMPLATE_DIR"
    rm -f "$TEMP_FILE"
    
    echo "✓ Templates installed"
}

main() {
    if check_templates; then exit 0; fi
    
    if install_templates_manual; then
        exit 0
    fi
    
    echo ""
    echo "⚠️ Download failed. Godot export templates (~1GB) download is slow from China."
    echo ""
    echo "Solutions:"
    echo "  1. Use VPN/proxy and retry: ./scripts/install-godot-templates.sh"
    echo ""
    echo "  2. Manual download with GitHub proxy:"
    echo "     curl -LO https://ghproxy.com/https://github.com/godotengine/godot/releases/download/${GODOT_VERSION}-stable/Godot_v${GODOT_VERSION}-stable_export_templates.tpz"
    echo "     unzip Godot_v${GODOT_VERSION}-stable_export_templates.tpz -d $TEMPLATE_DIR"
    echo ""
    echo "  3. Use pre-built web files (SKIP templates entirely):"
    echo "     # If you already have web export files, just copy them:"
    echo "     mkdir -p server/cmd/server/web"
    echo "     cp -r /your/web/files/* server/cmd/server/web/"
    echo "     make build && make run  # No need for make web"
    echo ""
    
    exit 1
}

main "$@"
