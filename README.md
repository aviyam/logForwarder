# Log Forwarder in Go

A lightweight, containerized Go application that acts as a log collector and forwarder. It receives logs from **Filebeat** (Lumberjack v2) and **Fluentd/Fluent Bit** (Forward protocol) and writes them directly to `stdout` as JSON.

This is ideal for containerized environments where you want to consolidate logs from various sidecars or agents and have them managed by the container runtime's logging driver.

## Features

- **Lumberjack v2 Protocol Support**: Receives logs from Filebeat on port `5044`.
- **Fluent Forward Protocol Support**: Receives logs from Fluentd or Fluent Bit on port `24224`.
  - Supports `Message`, `Forward`, and `PackedForward` modes.
  - Supports **Gzip compression** for `CompressedPackedForward`.
- **JSON Output**: All incoming logs are normalized and printed to `stdout` as JSON.
- **Graceful Shutdown**: Handles `SIGTERM` and `SIGINT` to ensure all active connections are closed properly.
- **Container Optimized**: Small footprint using a multi-stage Docker build.

## Getting Started

### Prerequisites

- [Docker](https://www.docker.com/) and [Docker Compose](https://docs.docker.com/compose/)
- [Go 1.24+](https://golang.org/) (if building locally)

### Running with Docker

1. **Build the image**:
   ```bash
   docker build -t logforwarder .
   ```

2. **Run the container**:
   ```bash
   docker run -p 5044:5044 -p 24224:24224 logforwarder
   ```

### Configuration

Configuration is handled via environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `LUMBERJACK_ADDR` | TCP address to listen for Filebeat logs | `:5044` |
| `FLUENT_ADDR` | TCP address to listen for Fluentd logs | `:24224` |

## Testing the Setup

A `docker-compose.yaml` and test configurations are provided to verify the application.

1. **Start the test environment**:
   ```bash
   docker compose up -d
   ```

2. **Test Filebeat**:
   Append a line to the watched log file:
   ```bash
   echo "{\"message\": \"test log from filebeat\", \"level\": \"info\"}" >> test/test.log
   ```

3. **Test Fluentd**:
   Send a log via Fluentd's HTTP input:
   ```bash
   curl -X POST -d 'json={"message":"test log from fluentd", "status": 200}' http://localhost:8888/test.tag
   ```

4. **Verify Output**:
   Check the `logforwarder` container logs:
   ```bash
   docker compose logs -f logforwarder
   ```

## Local Development

Using the included `Makefile`:

- `make build`: Compiles the binary locally.
- `make run`: Runs the application locally.
- `make test-up`: Starts the Docker Compose test environment.
- `make clean`: Removes binaries and stops test containers.

## License

MIT
