#!/usr/bin/env bash
set -euo pipefail

REPO="mr-coder20/FireScan"
VERSION="${1:-latest}"
INSTALL_DIR="${2:-/usr/local/bin}"

echo "🔥 FireScan Installer"
echo "====================="

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
    x86_64)  ARCH="amd64" ;;
    aarch64) ARCH="arm64" ;;
    armv7l)  ARCH="armv7" ;;
    *)       echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

case "$OS" in
    linux)   ;;
    darwin)  OS="darwin" ;;
    *)       echo "Unsupported OS: $OS"; exit 1 ;;
esac

echo "• Detected: $OS/$ARCH"
echo "• Installing to: $INSTALL_DIR"

# Download URL
URL="https://github.com/$REPO/releases/download/$VERSION/firescan-${OS}-${ARCH}.tar.gz"

# Create temp directory
TMP_DIR=$(mktemp -d)
trap "rm -rf $TMP_DIR" EXIT

echo "• Downloading from $URL..."
if ! curl -sSL --fail "$URL" 2>/dev/null; then
    # If latest, try to get the actual latest version
    if [ "$VERSION" = "latest" ]; then
        echo "  Trying to fetch latest release..."
        URL="https://github.com/$REPO/releases/latest/download/firescan-${OS}-${ARCH}.tar.gz"
    fi
    curl -sSL --fail "$URL" | tar -xz -C "$TMP_DIR" || {
        echo "  ✗ Download failed. Trying 'go install' instead..."
        go install "github.com/$REPO@latest" 2>/dev/null && {
            echo "  ✓ Installed via 'go install'"
            "$GOPATH/bin/firescan" --version
            exit 0
        }
        echo "  ✗ Install failed. Please download manually from:"
        echo "    https://github.com/$REPO/releases"
        exit 1
    }
else
    curl -sSL "$URL" | tar -xz -C "$TMP_DIR"
fi

# Install binary
sudo mkdir -p "$INSTALL_DIR"
if [ -f "$TMP_DIR/firescan" ]; then
    sudo mv "$TMP_DIR/firescan" "$INSTALL_DIR/"
else
    # Try to find binary in extracted files
    BIN=$(find "$TMP_DIR" -name "firescan" -type f | head -1)
    if [ -n "$BIN" ]; then
        sudo mv "$BIN" "$INSTALL_DIR/"
    else
        echo "  ✗ Binary not found in archive"
        exit 1
    fi
fi
sudo chmod +x "$INSTALL_DIR/firescan"

# Verify
echo "• Verifying installation..."
"$INSTALL_DIR/firescan" --version || echo "  (run 'firescan --version' after adding to PATH)"

# Add to PATH if needed
if ! echo "$PATH" | grep -q "$INSTALL_DIR"; then
    echo "• Adding $INSTALL_DIR to PATH..."
    echo "export PATH=\"\$PATH:$INSTALL_DIR\"" >> "$HOME/.bashrc"
    echo "export PATH=\"\$PATH:$INSTALL_DIR\"" >> "$HOME/.zshrc" 2>/dev/null || true
    echo "  Please restart your terminal or run: export PATH=\"\$PATH:$INSTALL_DIR\""
fi

echo ""
echo "✅ FireScan installed successfully!"
echo "   Run 'firescan --help' to get started"