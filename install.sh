#!/bin/bash
set -e

# Outfitpicker installer script
REPO="dh85/outfitpicker"
BINARY="outfitpicker"

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
    x86_64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) echo "Unsupported architecture: $ARCH" && exit 1 ;;
esac

case $OS in
    linux)
        if command -v apt >/dev/null 2>&1; then
            # Ubuntu/Debian - use .deb package
            echo "Installing via .deb package..."
            LATEST=$(curl -s https://api.github.com/repos/$REPO/releases/latest | grep tag_name | cut -d '"' -f 4)
            DEB_URL="https://github.com/$REPO/releases/download/$LATEST/${BINARY}_${LATEST#v}_linux_${ARCH}.deb"
            
            curl -sL "$DEB_URL" -o /tmp/outfitpicker.deb
            sudo dpkg -i /tmp/outfitpicker.deb
            rm /tmp/outfitpicker.deb
            
        elif command -v yum >/dev/null 2>&1 || command -v dnf >/dev/null 2>&1; then
            # RHEL/CentOS/Fedora - use .rpm package
            echo "Installing via .rpm package..."
            LATEST=$(curl -s https://api.github.com/repos/$REPO/releases/latest | grep tag_name | cut -d '"' -f 4)
            RPM_URL="https://github.com/$REPO/releases/download/$LATEST/${BINARY}_${LATEST#v}_linux_${ARCH}.rpm"
            
            curl -sL "$RPM_URL" -o /tmp/outfitpicker.rpm
            if command -v dnf >/dev/null 2>&1; then
                sudo dnf install -y /tmp/outfitpicker.rpm
            else
                sudo yum install -y /tmp/outfitpicker.rpm
            fi
            rm /tmp/outfitpicker.rpm
            
        else
            # Generic Linux - extract binary
            echo "Installing binary to /usr/local/bin..."
            LATEST=$(curl -s https://api.github.com/repos/$REPO/releases/latest | grep tag_name | cut -d '"' -f 4)
            TAR_URL="https://github.com/$REPO/releases/download/$LATEST/${BINARY}_Linux_${ARCH}.tar.gz"
            
            curl -sL "$TAR_URL" | sudo tar -xz -C /usr/local/bin $BINARY ${BINARY}-admin
        fi
        ;;
    darwin)
        # macOS
        if command -v brew >/dev/null 2>&1; then
            echo "Installing via Homebrew..."
            brew tap dh85/tap
            brew install outfitpicker
        else
            echo "Installing binary to /usr/local/bin..."
            LATEST=$(curl -s https://api.github.com/repos/$REPO/releases/latest | grep tag_name | cut -d '"' -f 4)
            TAR_URL="https://github.com/$REPO/releases/download/$LATEST/${BINARY}_Darwin_${ARCH}.tar.gz"
            
            curl -sL "$TAR_URL" | sudo tar -xz -C /usr/local/bin $BINARY ${BINARY}-admin
        fi
        ;;
    *)
        echo "Unsupported OS: $OS"
        exit 1
        ;;
esac

echo "✅ outfitpicker installed successfully!"
echo "Run 'outfitpicker --help' to get started."