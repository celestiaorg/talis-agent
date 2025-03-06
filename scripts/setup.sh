#!/bin/bash

# Create required directories
sudo mkdir -p /etc/talis-agent/payload

# Create default configuration file if it doesn't exist
if [ ! -f /etc/talis-agent/config.yaml ]; then
    sudo tee /etc/talis-agent/config.yaml > /dev/null << 'EOL'
http:
  port: 25550
logging:
  level: info
EOL
fi

# Set appropriate permissions
sudo chown -R $USER:$USER /etc/talis-agent
sudo chmod -R 755 /etc/talis-agent

echo "Setup completed successfully!" 