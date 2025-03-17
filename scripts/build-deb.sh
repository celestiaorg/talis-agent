#!/bin/bash
set -e

# Package details
PACKAGE_NAME="talis-agent"
VERSION=$(git describe --tags --always || echo "0.1.0")
ARCH="amd64"
MAINTAINER="DevOps Team"
DESCRIPTION="System metrics monitoring agent"

# Create package directory structure
PACKAGE_DIR="dist/debian/${PACKAGE_NAME}"
mkdir -p "${PACKAGE_DIR}/DEBIAN"
mkdir -p "${PACKAGE_DIR}/usr/local/bin"
mkdir -p "${PACKAGE_DIR}/etc/talis-agent"
mkdir -p "${PACKAGE_DIR}/lib/systemd/system"

# Copy binary
cp "bin/${PACKAGE_NAME}" "${PACKAGE_DIR}/usr/local/bin/"

# Copy config file
cp "configs/config.yaml" "${PACKAGE_DIR}/etc/talis-agent/"

# Create systemd service file
cat > "${PACKAGE_DIR}/lib/systemd/system/talis-agent.service" << EOL
[Unit]
Description=Talis Agent Service
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/talis-agent
Restart=always
User=root

[Install]
WantedBy=multi-user.target
EOL

# Create control file
cat > "${PACKAGE_DIR}/DEBIAN/control" << EOL
Package: ${PACKAGE_NAME}
Version: ${VERSION}
Architecture: ${ARCH}
Maintainer: ${MAINTAINER}
Description: ${DESCRIPTION}
 A Go-based service that monitors key system metrics (CPU, memory, disk, I/O, network)
 via the Prometheus client. Exposes HTTP endpoints for metrics collection and system
 interaction.
Section: utils
Priority: optional
EOL

# Create postinst script
cat > "${PACKAGE_DIR}/DEBIAN/postinst" << EOL
#!/bin/bash
set -e

# Create required directories
mkdir -p /etc/talis-agent/payload

# Set permissions
chown -R root:root /etc/talis-agent
chmod -R 755 /etc/talis-agent

# Enable and start service
systemctl daemon-reload
systemctl enable talis-agent
systemctl start talis-agent

exit 0
EOL

# Create prerm script
cat > "${PACKAGE_DIR}/DEBIAN/prerm" << EOL
#!/bin/bash
set -e

# Stop and disable service
systemctl stop talis-agent || true
systemctl disable talis-agent || true

exit 0
EOL

# Set permissions
chmod 755 "${PACKAGE_DIR}/DEBIAN/postinst"
chmod 755 "${PACKAGE_DIR}/DEBIAN/prerm"

# Build the package
dpkg-deb --build "${PACKAGE_DIR}" "dist/${PACKAGE_NAME}_${VERSION}_${ARCH}.deb"

echo "Package created: dist/${PACKAGE_NAME}_${VERSION}_${ARCH}.deb" 