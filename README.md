# Talis Agent

Talis Agent is a Go-based service designed to run on Linux servers, providing real-time system metrics collection and monitoring capabilities. It exposes several HTTP endpoints for accessing system information and executing commands.

## Features

- **System Metrics Collection**
  - CPU usage (total and per-core)
  - Memory utilization
  - Disk usage and I/O statistics
  - Network interface information and I/O counters
  - Host system information

- **HTTP Endpoints**
  - `/metrics`: Exposes system metrics in Prometheus-compatible format
  - `/alive`: Health check endpoint
  - `/ip`: Returns public IP addresses
  - `/payload`: Accepts POST data for storage
  - `/commands`: Executes system commands

## Requirements

- Linux operating system
- Go 1.20 or later
- System with sufficient permissions to access system metrics

## Installation

### Using Debian Package

1. Download the latest `.deb` package from the releases page
2. Install using dpkg:
   ```bash
   sudo dpkg -i talis-agent_<version>.deb
   ```

### Building from Source

1. Clone the repository:
   ```bash
   git clone https://github.com/celestiaorg/talis-agent.git
   cd talis-agent
   ```

2. Build the project:
   ```bash
   make build
   ```

3. Install the service:
   ```bash
   sudo make install
   ```

## Configuration

The agent uses a configuration file located at `/etc/talis-agent/config.yaml`. A sample configuration is provided in the `configs` directory.

Example configuration:
```yaml
http:
  port: 25550  # Default HTTP port
logging:
  level: info  # Log level (debug, info, warn, error)
```

## Usage

### Starting the Service

```bash
sudo systemctl start talis-agent
```

### Checking Service Status

```bash
sudo systemctl status talis-agent
```

### Accessing Endpoints

1. **Metrics Endpoint**
   ```bash
   curl http://localhost:25550/metrics
   ```

2. **Health Check**
   ```bash
   curl http://localhost:25550/alive
   ```

3. **IP Information**
   ```bash
   curl http://localhost:25550/ip
   ```

4. **Payload Endpoint**
   ```bash
   curl -X POST -d "test data" http://localhost:25550/payload
   ```

5. **Commands Endpoint**
   ```bash
   curl -X POST -d "ls -la" http://localhost:25550/commands
   ```

## Development

### Project Structure

```
talis-agent/
├── cmd/           # Main application entrypoint
├── internal/      # Internal packages
│   ├── metrics/   # System metrics collection
│   └── logging/   # Logging utilities
├── pkg/           # Shared libraries
├── configs/       # Configuration files
├── scripts/       # Build and deployment scripts
└── tests/         # Test files
```

### Building for Development

```bash
make dev
```

### Running Tests

```bash
make test
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

For support, please open an issue in the GitHub repository or contact the maintainers. 
