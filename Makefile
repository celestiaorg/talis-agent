.PHONY: build test clean package

# Build the binary
build:
	go build -o talis-agent

# Run tests
test:
	go test ./...

# Clean build artifacts
clean:
	rm -f talis-agent
	rm -rf dist/

# Create deb package
package: build
	mkdir -p dist/DEBIAN
	mkdir -p dist/etc/talis-agent
	mkdir -p dist/usr/local/bin

	# Copy binary
	cp talis-agent dist/usr/local/bin/

	# Copy config file
	cp /etc/talis-agent/config.yaml dist/etc/talis-agent/

	# Create control file
	cat > dist/DEBIAN/control << 'EOL'
Package: talis-agent
Version: 1.0.0
Section: utils
Priority: optional
Architecture: amd64
Maintainer: Celestia <info@celestia.org>
Description: Talis Agent - System Metrics Collection Service
 A Go-based service that monitors system metrics and exposes them via HTTP endpoints.
EOL

	# Create postinst script
	cat > dist/DEBIAN/postinst << 'EOL'
#!/bin/bash
set -e

# Create required directories
mkdir -p /etc/talis-agent/payload

# Set permissions
chmod 755 /usr/local/bin/talis-agent
chmod 755 /etc/talis-agent
chmod 755 /etc/talis-agent/payload
EOL

	chmod 755 dist/DEBIAN/postinst

	# Build the package
	dpkg-deb --build dist talis-agent_1.0.0_amd64.deb 