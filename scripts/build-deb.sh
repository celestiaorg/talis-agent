#!/bin/bash
set -e

# Script configuration
PROJECT_NAME="talis-agent"
VERSION="0.1.0"
ARCH=$(dpkg --print-architecture)
BUILD_DIR="build"
PACKAGE_DIR="${BUILD_DIR}/${PROJECT_NAME}_${VERSION}_${ARCH}"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Print step with color
print_step() {
    echo -e "${GREEN}==> $1${NC}"
}

# Print error with color
print_error() {
    echo -e "${RED}ERROR: $1${NC}"
}

# Check for required tools
check_dependencies() {
    print_step "Checking dependencies..."
    
    local DEPS=(dpkg-deb fakeroot dpkg-dev debhelper golang-go)
    local MISSING=()
    
    for dep in "${DEPS[@]}"; do
        if ! command -v "$dep" >/dev/null 2>&1; then
            MISSING+=("$dep")
        fi
    done
    
    if [ ${#MISSING[@]} -ne 0 ]; then
        print_error "Missing required tools: ${MISSING[*]}"
        echo "Please install them with: sudo apt-get install ${MISSING[*]}"
        exit 1
    fi
}

# Clean build directory
clean_build() {
    print_step "Cleaning build directory..."
    rm -rf "$BUILD_DIR"
    mkdir -p "$PACKAGE_DIR"
}

# Build Go binary
build_binary() {
    print_step "Building Go binary..."
    go build -o "$PACKAGE_DIR/usr/bin/$PROJECT_NAME" cmd/agent/main.go
}

# Create package structure
create_package_structure() {
    print_step "Creating package structure..."
    
    # Create directories
    mkdir -p "$PACKAGE_DIR/DEBIAN"
    mkdir -p "$PACKAGE_DIR/usr/bin"
    mkdir -p "$PACKAGE_DIR/etc/$PROJECT_NAME"
    mkdir -p "$PACKAGE_DIR/lib/systemd/system"
    mkdir -p "$PACKAGE_DIR/var/lib/$PROJECT_NAME"
    mkdir -p "$PACKAGE_DIR/var/log/$PROJECT_NAME"
    
    # Copy files
    cp debian/control "$PACKAGE_DIR/DEBIAN/"
    cp debian/postinst "$PACKAGE_DIR/DEBIAN/"
    cp debian/postrm "$PACKAGE_DIR/DEBIAN/"
    cp config.yaml "$PACKAGE_DIR/etc/$PROJECT_NAME/"
    cp debian/talis-agent.service "$PACKAGE_DIR/lib/systemd/system/"
    
    # Set permissions
    chmod 755 "$PACKAGE_DIR/DEBIAN/postinst"
    chmod 755 "$PACKAGE_DIR/DEBIAN/postrm"
    chmod 644 "$PACKAGE_DIR/etc/$PROJECT_NAME/config.yaml"
    chmod 644 "$PACKAGE_DIR/lib/systemd/system/talis-agent.service"
}

# Build Debian package
build_package() {
    print_step "Building Debian package..."
    fakeroot dpkg-deb --build "$PACKAGE_DIR"
}

# Verify package
verify_package() {
    print_step "Verifying package..."
    local PACKAGE_FILE="${BUILD_DIR}/${PROJECT_NAME}_${VERSION}_${ARCH}.deb"
    
    if ! dpkg-deb -I "$PACKAGE_FILE" >/dev/null; then
        print_error "Package verification failed"
        exit 1
    fi
    
    echo -e "${GREEN}Package created successfully: ${PACKAGE_FILE}${NC}"
    echo "You can install it with: sudo dpkg -i $PACKAGE_FILE"
}

# Main execution
main() {
    check_dependencies
    clean_build
    build_binary
    create_package_structure
    build_package
    verify_package
}

# Run main function
main 