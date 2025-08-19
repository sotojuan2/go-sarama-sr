# Kafka Producer for Confluent Cloud

A high-performance, asynchronous Kafka producer built in Go that integrates with Confluent Cloud Schema Registry for producing Protobuf-serialized messages.

## Overview

This producer generates random `shoe` events and sends them to a Kafka topic using:
- **Sarama** for asynchronous Kafka production
- **Confluent Schema Registry** for Protobuf serialization
- **Random data generation** for realistic testing scenarios

## Quick Start

1. **Configure environment variables:**
```bash
cp .env.example .env
# Edit .env with your Confluent Cloud credentials
```

2. **Run the producer:**
```bash
go run main.go
```

3. **Stop gracefully:**
Press `Ctrl+C` to stop and see final statistics.

## Configuration

Set these environment variables (or use `.env` file):

```bash
# Kafka Configuration
KAFKA_BOOTSTRAP_SERVERS=your-cluster.confluent.cloud:9092
KAFKA_API_KEY=your-api-key
KAFKA_API_SECRET=your-api-secret
KAFKA_TOPIC=shoe-events

# Schema Registry
SCHEMA_REGISTRY_URL=https://your-registry.confluent.cloud
SCHEMA_REGISTRY_API_KEY=your-sr-key
SCHEMA_REGISTRY_API_SECRET=your-sr-secret

# Optional Settings
MESSAGE_INTERVAL=1s        # How often to send messages
LOG_LEVEL=info            # debug, info, warn, error
PRODUCER_TIMEOUT=30s      # Kafka producer timeout
```

## Architecture

```
┌─────────────────┐    ┌──────────────┐    ┌─────────────────┐
│  Data Generator │───▶│   Producer   │───▶│ Confluent Cloud │
│  (Random Shoes) │    │   (Sarama)   │    │     Kafka       │
└─────────────────┘    └──────────────┘    └─────────────────┘
                              │
                              ▼
                    ┌──────────────────┐
                    │ Schema Registry  │
                    │   (Protobuf)     │
                    └──────────────────┘
```

## Shoe Schema

The producer generates random shoe events based on this Protobuf schema:

```proto
syntax = "proto3";
package main;
option go_package = "./pb";

message Shoe {
  int64 id = 1;
  string brand = 2;
  string name = 3;
  double sale_price = 4;
  double rating = 5;
}
```

## Performance

The producer is optimized for high throughput:
- Asynchronous message production
- Configurable batching and compression
- Proper error handling and recovery
- Graceful shutdown with message flushing

Expected performance: **1000+ messages/second** depending on network and Kafka cluster configuration.

## Development

```bash
# Generate Protobuf code
protoc --go_out=. --go_opt=paths=source_relative shoe.proto

# Run tests
go test ./...

# Build binary
go build -o kafka-producer main.go
```

## Docker

```bash
# Build image
docker build -t kafka-producer .

# Run container
docker run --env-file .env kafka-producer
```

## Monitoring

The producer outputs structured logs showing:
- Connection status
- Message production rate
- Success/error counts
- Schema registration status
- Graceful shutdown progress

Example output:
```
2025/08/19 10:30:15 🚀 Starting Kafka Producer...
2025/08/19 10:30:16 ✅ Connected to Kafka cluster
2025/08/19 10:30:16 ✅ Schema Registry connected
2025/08/19 10:30:17 📊 Rate: 245.2 msg/s | Produced: 1226 | Errors: 0
```

## License

MIT
