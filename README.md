# Kafka Consumer for AI Image Commands

Go-based Kafka consumer service that processes AI image commands with real image processing capabilities.

## Features

- **Kafka Consumer/Producer**: Consume and produce image processing commands
- **Real Image Processing**: Resize, crop, filter, transform images using `disintegration/imaging`
- **Protobuf & JSON Support**: Configurable message format
- **Structured Logging**: JSON logs via `log/slog`
- **Prometheus Metrics**: `/metrics` endpoint for monitoring
- **OpenTelemetry Tracing**: Distributed tracing support
- **Chi Router**: Fast and lightweight HTTP router
- **Validation**: Request validation with `go-playground/validator`
- **Custom Errors**: Typed error handling with JSON responses
- **Clean Architecture**: delivery → usecase → repository layers

## Quick Start

### Prerequisites

- Go 1.21+
- Docker & Docker Compose
- Make

### Run with Docker (recommended)

```bash
# Clone the repository
git clone https://github.com/F-i-i-X-i-i/kafka-consumer-it.git
cd kafka-consumer-it

# Start all services (Kafka, Zookeeper, Consumer, Prometheus, Grafana)
make docker-up

# Check health
make health

# Send test message
make send-test

# View logs
make docker-logs

# Stop all services
make docker-down
```

### Run locally (development)

```bash
# Install dependencies
go mod download

# Start Kafka (required)
docker-compose up -d kafka zookeeper

# Run the consumer
make run

# Or with real image processing
PROCESSOR_MODE=real make run
```

## Configuration

| Environment Variable | Default | Description |
|---------------------|---------|-------------|
| `KAFKA_BROKERS` | `localhost:9092` | Kafka broker addresses |
| `KAFKA_TOPIC` | `image-commands` | Topic to consume from |
| `KAFKA_GROUP_ID` | `image-processor-group` | Consumer group ID |
| `PROCESSOR_MODE` | `stub` | `stub` or `real` |
| `MESSAGE_FORMAT` | `json` | `json` or `protobuf` |
| `OUTPUT_DIR` | `/tmp/processed-images` | Output directory for processed images |
| `HTTP_PORT` | `8080` | HTTP server port |
| `LOG_LEVEL` | `info` | Log level (debug/info/warn/error) |

You can also use a `config.yaml` file:

```yaml
kafka_brokers: "kafka:29092"
kafka_topic: "image-commands"
processor_mode: "real"
log_level: "debug"
```

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/health` | GET | Health check |
| `/ready` | GET | Readiness check |
| `/stats` | GET | Processing statistics |
| `/metrics` | GET | Prometheus metrics |
| `/send` | POST | Send image command |

## Supported Commands

| Command | Description | Parameters |
|---------|-------------|------------|
| `resize` | Resize image | `width`, `height` |
| `crop` | Crop image region | `x`, `y`, `width`, `height` |
| `filter` | Apply filter | `filter_type` (blur/sharpen/grayscale/invert), `intensity` |
| `transform` | Transform image | `rotation_degrees`, `flip_horizontal`, `flip_vertical` |
| `analyze` | AI image analysis | `models` |
| `remove_background` | Remove background | `output_format`, `high_quality` |

## Example Request

```bash
curl -X POST http://localhost:8080/send \
  -H "Content-Type: application/json" \
  -d '{
    "id": "cmd-123",
    "command": "resize",
    "image_url": "https://example.com/image.jpg",
    "parameters": {
      "width": 800,
      "height": 600
    }
  }'
```

## Make Commands

```bash
make build        # Build binary
make run          # Run locally
make test         # Run tests
make lint         # Run linters
make docker-up    # Start docker-compose
make docker-down  # Stop docker-compose
make docker-logs  # View logs
make proto        # Generate protobuf code
make health       # Check health endpoint
make send-test    # Send test message
```

## Project Structure

```
├── cmd/consumer/main.go          # Entry point
├── internal/
│   ├── app/                      # Application lifecycle
│   ├── config/                   # Viper configuration
│   ├── delivery/
│   │   ├── api/                  # HTTP handlers (chi router)
│   │   └── queue/                # Kafka consumer handler
│   ├── kafka/                    # Kafka producer
│   ├── models/api/               # API DTOs
│   ├── pkg/
│   │   ├── customerrors/         # Custom error types
│   │   ├── logger/               # Structured logging
│   │   ├── metrics/              # Prometheus metrics
│   │   └── tracing/              # OpenTelemetry
│   ├── processor/                # Command processing
│   ├── repository/storage/       # S3/local storage interface
│   └── usecase/                  # Business logic
├── proto/                        # Protobuf definitions
├── postman/                      # Postman collection
├── Makefile
├── Dockerfile
├── docker-compose.yml
└── prometheus.yml
```

## Monitoring

- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000 (admin/admin)

### Available Metrics

- `kafka_consumer_messages_processed_total` - Total messages processed
- `kafka_consumer_message_processing_duration_seconds` - Processing duration
- `kafka_consumer_http_requests_total` - HTTP requests count

## Testing

```bash
# Run unit tests
make test

# Run tests with coverage
go test -cover ./...

# Import Postman collection for API testing
# File: postman/kafka-consumer-tests.postman_collection.json
```

## Development

```bash
# Generate protobuf code
make proto

# Build for Linux
make build

# Build Docker image
docker build -t kafka-consumer .
```

