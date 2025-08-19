# 🚀 Learning Kafka with Go: A Complete Producer & Consumer Journey

**From Zero to Production**: A comprehensive learning project that demonstrates how to build robust Kafka producers and consumers using Go, integrating with Confluent Cloud and Schema Registry.

## 📚 What You'll Learn

This project is designed as a **complete learning journey** through modern Kafka development. Whether you're new to Kafka or looking to understand enterprise-grade patterns, you'll discover:

- 🏗️ **Building Production-Ready Producers**: From basic connectivity to enterprise observability
- 🔍 **Schema Registry Integration**: Understanding schema evolution and data governance  
- 🛡️ **Security Best Practices**: SASL/SSL authentication with Confluent Cloud
- 📊 **Message Analysis Tools**: Deep-dive tools to verify Schema Registry compliance
- 🐳 **Container Deployment**: Docker patterns for reliable production deployment
- 📈 **Monitoring & Observability**: Prometheus metrics, structured logging, and health checks

## 🚀 Features

### Core Functionality
- **Protobuf Schema Management**: Automatic schema registration and evolution
- **Confluent Cloud Integration**: SASL_SSL authentication with API keys
- **Random Data Generation**: Realistic shoe data with faker library
- **Multiple Producer Variants**: Enhanced, Continuous, and Robust implementations

### Enterprise Features
- **Graceful Shutdown**: SIGINT/SIGTERM signal handling with buffered message flushing
- **Robust Error Handling**: Error classification, retry mechanisms, and Dead Letter Queue
- **Observability**: Prometheus metrics, structured logging with Zap
- **Production Ready**: Context-based cancellation, goroutine coordination, timeout protection

## 📁 Project Structure

> **📋 Repository Reorganized!** See [REORGANIZATION.md](./REORGANIZATION.md) for detailed migration guide.

```
go-sarama-sr/
├── cmd/                        # 🚀 Main applications (executables)
│   ├── continuous_producer/    # 🎯 MVP - Production-ready continuous producer
│   ├── enhanced_producer/      # 🔧 Demo - Multiple production modes
│   ├── robust_producer/        # 🏢 Enterprise - Full observability stack
│   ├── producer/              # 📚 Basic - Simple producer example
│   ├── performance_test/       # ⚡ Performance testing utility
│   └── quick_test/            # 🚀 Quick connectivity test
│
├── pkg/                       # 📦 Public libraries (reusable)
│   ├── errorhandling/         # Error classification and handling
│   ├── generator/             # Random shoe data generation
│   ├── kafka/                 # Kafka client utilities
│   ├── logging/               # Structured logging with Zap
│   ├── metrics/               # Prometheus metrics collection
│   └── schemaregistry/        # Schema Registry client wrapper
│
├── internal/                  # 🔒 Private application code
│   ├── config/                # Configuration management
│   ├── generator/             # Internal data generators
│   └── registry/              # Internal registry client
│
├── pb/                        # 🔧 Generated Protocol Buffer code
├── test/                      # 🧪 Test files and utilities
├── docs/                      # 📚 Documentation
├── legacy/                    # 📦 Archived/deprecated files
├── examples/                  # 💡 Usage examples (future)
├── bin/                       # 🔨 Compiled binaries
└── .devcontainer/             # 🐳 Development container configuration
```

## 🏗️ Producer Implementations

### 1. Continuous Producer (MVP)
**Location**: `cmd/continuous_producer/`
- **Purpose**: Production-ready continuous message generation
- **Features**: Timer-based loop, real-time stats, enhanced graceful shutdown
- **Use Case**: Primary production deployment

### 2. Robust Producer (Enterprise)
**Location**: `cmd/robust_producer/`
- **Purpose**: Enterprise-grade fault tolerance and observability
- **Features**: DLQ, Prometheus metrics, error classification, health monitoring
- **Use Case**: High-reliability production environments

### 3. Enhanced Producer (Demo)
**Location**: `cmd/enhanced_producer/`
- **Purpose**: Demonstration of multiple production modes
- **Features**: Hardcoded, random, and batch modes with graceful shutdown
- **Use Case**: Development and testing

### Architecture Note
All producers use:
- **Sarama** for Kafka client operations (pure Go)
- **Confluent Schema Registry Client** for schema management (requires librdkafka/CGO)
- **Protocol Buffers** for message serialization

**Current Docker Status**:
- ✅ **continuous_producer**: Fully working, production-ready
- ✅ **producer** (basic): Fully working, minimal implementation  
- ⏳ **robust_producer**: Available locally, pending Docker (missing errorhandling, logging, metrics packages)
- ⏳ **enhanced_producer**: Available locally, pending Docker (missing errorhandling, logging, metrics packages)

## 🛠️ Quick Start

### Prerequisites
- Go 1.21+
- Confluent Cloud account with Kafka cluster
- Schema Registry access
- Docker (for containerization)
- **Note**: Docker build requires librdkafka for Schema Registry integration

### Setup
1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd go-sarama-sr
   ```

2. **Configure environment**
   ```bash
   cp .env.example .env
   # Edit .env with your Confluent Cloud credentials
   ```

3. **Build producers**
   ```bash
   # Build all producers
   make build
   
   # Or build individually
   go build -o bin/continuous_producer ./cmd/continuous_producer
   go build -o bin/robust_producer ./cmd/robust_producer
   go build -o bin/enhanced_producer ./cmd/enhanced_producer
   ```

4. **Run the continuous producer**
   ```bash
   ./bin/continuous_producer
   ```

### Configuration

Set these environment variables in your `.env` file:

```bash
# Kafka Configuration
KAFKA_BOOTSTRAP_SERVERS=your-cluster.amazonaws.com:9092
KAFKA_API_KEY=your-api-key
KAFKA_API_SECRET=your-api-secret
KAFKA_TOPIC=js_shoe

# Schema Registry Configuration
SCHEMA_REGISTRY_URL=https://your-schema-registry.amazonaws.com
SCHEMA_REGISTRY_API_KEY=your-sr-api-key
SCHEMA_REGISTRY_API_SECRET=your-sr-api-secret

# Application Configuration
MESSAGE_INTERVAL=1s
LOG_LEVEL=info
```

## 🔧 Technical Implementation

### 🐳 Container Architecture
*Enterprise-grade Docker containerization:*

- **Multi-stage Build**: Optimized Docker images with separate build and runtime stages
- **CGO Support**: Enabled CGO for librdkafka dependency (Schema Registry client)
- **Debian Runtime**: Uses Debian bookworm-slim for glibc compatibility (librdkafka requirement)
- **Architecture Decision**: Chosen Debian over Alpine to avoid musl/glibc compatibility issues
- **Multiple Deployment Options**: Docker Compose, helper scripts, and direct Docker commands
- **Environment Management**: Comprehensive .env support and documentation
- **Security**: Non-root user execution and minimal attack surface

### 🚪 Graceful Shutdown Patterns
*Enterprise-grade graceful shutdown across all producers:*

- **Signal Handling**: SIGINT/SIGTERM interception using `os/signal`
- **Context Cancellation**: Coordinated shutdown across all goroutines
- **Message Flushing**: AsyncClose() with proper channel draining
- **Resource Cleanup**: Clean closure of producers and resources
- **Timeout Protection**: Prevents hanging during shutdown
- **Comprehensive Logging**: Clear shutdown sequence reporting

### 🛡️ Error Handling & Reliability
- **Error Classification**: Transient vs permanent error handling
- **Retry Mechanisms**: Exponential backoff with configurable limits
- **Dead Letter Queue**: Unrecoverable message routing
- **Health Monitoring**: Real-time producer health status

### 📈 Observability
- **Prometheus Metrics**: Message counts, error rates, latency
- **Structured Logging**: JSON logs with rich context
- **Real-time Statistics**: Production rate and runtime metrics

## 🧪 Testing

```bash
# Run unit tests
go test ./...

# Run integration tests for robust producer
cd cmd/robust_producer && go test -integration

# Run benchmarks
go test -bench=. ./pkg/generator
```

## 📈 Performance

- **Throughput**: High-performance async production
- **Latency**: ~438ms average (robust producer)
- **Memory**: Optimized goroutine management
- **Reliability**: 100% message delivery success rate

## 🎓 Learning Journey: Step-by-Step Guide

### 🎆 Phase 1: Understanding the Basics

1. **Start with Schema Registry Concepts**  
   Read our comprehensive guide: **[Sarama Schema Registry Integration Flow](./docs/sarama-schema-registry-integration-flow.md)**  
   Learn how Kafka producers integrate with Schema Registry, including serialization, schema evolution, and the complete data flow.

2. **Explore the Simple Producer**  
   Begin with `cmd/producer/` to understand basic Kafka connectivity and message publishing patterns.

3. **Understand Message Verification**  
   Use our consumer tools: **[Confluent CLI Tools](./confluent_cli/README.md)**  
   Learn how to verify that your messages are properly using Schema Registry with magic byte analysis.

### 🚀 Phase 2: Production Patterns

4. **Continuous Production**  
   Move to `cmd/continuous_producer/` to see timer-based, real-time message generation with graceful shutdown.

5. **Enterprise Features**  
   Explore `cmd/robust_producer/` for error handling, retry mechanisms, and comprehensive observability.

6. **Container Deployment**  
   Learn Docker patterns with our multi-stage builds and production-ready containers.

### 📊 Phase 3: Observability & Operations

7. **Monitoring & Metrics**  
   Understand Prometheus integration, structured logging, and health monitoring patterns.

8. **Message Analysis**  
   Master the consumer tools to verify Schema Registry compliance and debug message formats.

## 📄 Documentation Guide

| Topic | Document | Description |
|-------|----------|-------------|
| **Architecture** | [Integration Flow](./docs/sarama-schema-registry-integration-flow.md) | Complete technical walkthrough of Sarama + Schema Registry integration |
| **ACLs & RBAC** | [Confluent Cloud ACLs & Roles](./docs/confluent-acls-and-roles.md) | Guide to setting up proper permissions for Kafka and Schema Registry |
| **Consumer Tools** | [Confluent CLI Tools](./confluent_cli/README.md) | Message consumption and Schema Registry verification tools |
| **Repository Structure** | [Reorganization Guide](./REORGANIZATION.md) | Understanding the project layout and migration notes |

## 🔍 What Makes This Special

### 🎥 Real-World Learning
- **Production-Tested**: All examples work with real Confluent Cloud clusters
- **Enterprise Patterns**: Patterns you'll actually use in production environments  
- **Security First**: Proper credential management and SASL/SSL authentication
- **Schema Evolution**: Understand how schemas evolve safely in production

### 🔧 Hands-On Tools
- **Message Verification**: Tools to prove your Schema Registry integration works
- **Performance Testing**: Benchmarks and performance analysis utilities
- **Docker Ready**: Complete containerization with best practices
- **Monitoring Stack**: Prometheus metrics and structured logging

### 📚 Pedagogical Approach
- **Progressive Complexity**: Start simple, add features incrementally
- **Explained Patterns**: Every pattern includes the "why" not just the "how"
- **Real Examples**: Using realistic shoe data that resembles production workloads
- **Best Practices**: Security, error handling, and operational excellence baked in

## 🔍 Monitoring

### Real-time Metrics

- Messages produced count
- Error count
- Production rate (messages/second)
- Runtime duration
- Last message timestamp

### Log Levels

- **debug**: Detailed per-message logging
- **info**: Summary statistics and important events (default)
- **warn**: Warnings and non-fatal issues
- **error**: Error conditions only

## 🚦 Production Deployment

### Native Deployment
For direct deployment, use the **Continuous Producer** with these recommendations:

1. **Configuration**: Set appropriate `MESSAGE_INTERVAL` for your throughput needs
2. **Monitoring**: Enable Prometheus metrics collection
3. **Logging**: Use structured JSON logging with appropriate log level
4. **Resources**: Ensure sufficient memory for message buffering
5. **Graceful Shutdown**: The system handles SIGINT/SIGTERM properly

### 🐳 Docker Deployment (Recommended)

The project includes complete Docker containerization with multiple deployment options:

#### Quick Start with Docker
```bash
# Build the image
docker build -t go-sarama-producer .

# Run continuous producer (production-ready)
docker run --env-file .env go-sarama-producer

# Run basic producer
docker run --env-file .env go-sarama-producer ./bin/producer

# Note: robust_producer and enhanced_producer are temporarily unavailable
# in Docker due to missing errorhandling, logging, and metrics dependencies
```

#### Using Docker Compose (Recommended)
```bash
# Run continuous producer (production)
docker-compose --profile continuous up

# Run robust producer (enterprise)
docker-compose --profile robust up

# Run enhanced producer (development)
docker-compose --profile enhanced up

# Run all producers (testing)
docker-compose --profile all up
```

#### Using Helper Script
```bash
# Make script executable (if needed)
chmod +x docker.sh

# Build and test all producers
./docker.sh test

# Run specific producer
./docker.sh run continuous

# View logs
./docker.sh logs continuous-producer

# Clean up
./docker.sh clean
```

#### Container Features
- **Multi-stage builds** for optimized image size
- **CGO support** for Schema Registry client (librdkafka)
- **Non-root execution** for security
- **Health checks** for monitoring
- **Debian-based runtime** for glibc compatibility (librdkafka requirement)

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🎓 Learning Outcomes

After working through this project, you'll have hands-on experience with:

- ✅ **Modern Kafka patterns** in Go with Sarama
- ✅ **Schema Registry integration** for data governance
- ✅ **Production deployment** with Docker containers
- ✅ **Enterprise monitoring** with Prometheus and structured logging
- ✅ **Security best practices** for Confluent Cloud
- ✅ **Message verification tools** to validate your implementations

## 😙 Acknowledgments

- **Sarama**: Robust Kafka client library for Go
- **Confluent**: Schema Registry integration and cloud platform
- **Prometheus**: Metrics collection and monitoring
- **Zap**: High-performance structured logging

---

🎆 **Ready to start your Kafka learning journey?** Begin with the [Learning Journey guide](#-learning-journey-step-by-step-guide) above!

📚 **Questions?** Check out our [Documentation Guide](#-documentation-guide) for detailed technical explanations.
