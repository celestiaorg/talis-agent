# Talis Agent

A system monitoring agent written in Go for Linux machines that continuously collects and reports system metrics to a central API server.

## Features

- Collects system metrics every 5 seconds:
  - CPU usage
  - Memory usage
  - Disk usage
  - I/O statistics
  - Network usage
- Sends regular check-ins to the API server
- Hosts an HTTP server on port 25550 for:
  - Receiving and storing payloads
  - Executing bash commands (with token authentication)
- Prometheus metrics integration
- Secure token-based authentication

## Requirements

- Go 1.18 or later
- Debian-based Linux distribution
- Access to the central API server

## Installation

### From Source

1. Clone the repository:
```bash
git clone https://github.com/celestiaorg/talis-agent.git
cd talis-agent
```

2. Build the agent:
```bash
go build -o talis-agent cmd/agent/main.go
```

3. Create a configuration file:
```bash
cp config.yaml /etc/talis-agent/config.yaml
```

4. Edit the configuration file with your settings:
```yaml
api_server: "http://your-api-server:8080"
token: "your-static-token"
checkin_interval: "5s"
http_port: 25550
log_level: "info"
metrics:
  collection_interval: "5s"
  endpoints:
    telemetry: "/v1/agent/telemetry"
    checkin: "/v1/agent/checkin"
payload:
  path: "/etc/talis-agent/payload"
```

### Using Debian Package (Coming Soon)

A Debian package will be available for easy installation on Debian-based systems.

## Usage

Run the agent with:

```bash
./talis-agent -config /path/to/config.yaml
```

### HTTP Endpoints

- `/payload` - POST endpoint for receiving payloads (requires token authentication)
- `/commands` - POST endpoint for executing bash commands (requires token authentication)
- `/metrics` - GET endpoint for Prometheus metrics

## Development

### Project Structure

- `cmd/agent/` - Main application entry point
- `internal/` - Internal packages
  - `config/` - Configuration handling
  - `metrics/` - System metrics collection
  - `http/` - HTTP server implementation
- `pkg/` - Reusable packages
- `tests/` - Integration and system tests
- `scripts/` - Build and packaging scripts

### Running Tests

```bash
go test ./...
```

### Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Security

- The agent uses a static token for authentication
- All sensitive endpoints require token authentication
- Command execution is restricted to authenticated requests only
- Payload storage is isolated to a specific directory

## Support

For support, please open an issue in the GitHub repository or contact the maintainers.
