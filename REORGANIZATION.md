# Repository Reorganization Guide

## 🔄 Overview

This document explains the reorganization of the `go-sarama-sr` repository to follow Go best practices and provide better structure for development and maintenance.

## 📁 New Repository Structure

```
go-sarama-sr/
├── cmd/                        # Main applications (executables)
│   ├── continuous_producer/    # 🎯 MVP - Production-ready continuous producer
│   ├── enhanced_producer/      # 🔧 Demo - Multiple production modes
│   ├── robust_producer/        # 🏢 Enterprise - Full observability stack
│   ├── producer/              # 📚 Basic - Simple producer example
│   ├── performance_test/       # ⚡ Performance testing utility
│   └── quick_test/            # 🚀 Quick connectivity test
│
├── pkg/                       # Public libraries (reusable)
│   ├── errorhandling/         # Error classification and handling
│   ├── generator/             # Random shoe data generation
│   ├── kafka/                 # Kafka client utilities
│   ├── logging/               # Structured logging with Zap
│   ├── metrics/               # Prometheus metrics collection
│   └── schemaregistry/        # Schema Registry client wrapper
│
├── internal/                  # Private application code
│   ├── config/                # Configuration management
│   ├── generator/             # Internal data generators
│   └── registry/              # Internal registry client
│
├── pb/                        # Generated Protocol Buffer code
│   ├── shoe.proto             # Protobuf schema definition
│   └── shoe.pb.go             # Generated Go structs (created in devcontainer)
│
├── test/                      # Test files and utilities
│   ├── config_test.go         # Configuration testing
│   ├── connectivity_validation_test.go  # Connection tests
│   ├── schemaregistry_test.go # Schema Registry tests
│   ├── shoe_test.go           # Protobuf message tests
│   ├── test_comprehensive_validation.go  # Full validation suite
│   ├── test_enhanced_auth.go  # Authentication tests
│   └── test_sr_connectivity.go # SR connectivity tests
│
├── docs/                      # Documentation
│   └── sarama-schema-registry-integration-flow.md  # Technical integration guide
│
├── legacy/                    # Archived/deprecated files
│   ├── main_clean.go          # Legacy main implementation
│   ├── main_new.go           # Experimental main version
│   ├── producer.go           # Old root producer
│   ├── producer_enhanced.go  # Duplicated enhanced producer
│   ├── producer_new.go       # Experimental producer
│   └── producer_simple.go    # Duplicated simple producer
│
├── examples/                  # Usage examples (future)
├── bin/                      # Compiled binaries
├── .devcontainer/            # Development container configuration
├── .taskmaster/              # Project management files
└── backup/                   # Development backups
```

## 🚀 Quick Start Guide

### For Development (Recommended)

```bash
# 1. Open in devcontainer (VS Code)
code .

# 2. Inside devcontainer, generate protobuf
protoc --go_out=./pb --go_opt=paths=source_relative pb/shoe.proto

# 3. Build the desired producer
go build -o bin/continuous_producer ./cmd/continuous_producer
go build -o bin/robust_producer ./cmd/robust_producer
go build -o bin/enhanced_producer ./cmd/enhanced_producer

# 4. Run with your .env configuration
./bin/continuous_producer
```

### For Quick Testing

```bash
# Test connectivity
go run ./cmd/quick_test

# Performance testing
go run ./cmd/performance_test
```

## 📋 Reorganization Changes

### ✅ What Was Moved

| Original Location | New Location | Purpose |
|------------------|--------------|---------|
| `main_new.go` | `legacy/main_new.go` | Experimental implementation |
| `producer.go` | `legacy/producer.go` | Duplicate of cmd/producer |
| `producer_enhanced.go` | `legacy/producer_enhanced.go` | Duplicate of cmd/enhanced_producer |
| `producer_simple.go` | `legacy/producer_simple.go` | Legacy simple implementation |
| `producer_new.go` | `legacy/producer_new.go` | Experimental producer |
| `test_*.go` | `test/test_*.go` | Test files organized |
| `shoe.proto` | `pb/shoe.proto` | Protocol Buffer definitions |
| Binaries | `bin/` (gitignored) | Compiled executables |

### ✅ What Was Cleaned

- **Duplicate files**: Removed redundant implementations
- **Compiled binaries**: Moved to `bin/` directory (gitignored)
- **Legacy code**: Archived in `legacy/` for reference
- **Root clutter**: Cleaned main directory

### ✅ What Was Preserved

- **All working implementations** in `cmd/`
- **All library code** in `pkg/` and `internal/`
- **All configuration files** (`.env.example`, `docker-compose.yml`, etc.)
- **All documentation** and task management files
- **Complete git history**

## 🎯 Application Guide

### 1. Continuous Producer (MVP - Production Ready)
**Location**: `cmd/continuous_producer/`
**Purpose**: Production-ready continuous message generation
**Features**: 
- Timer-based message production
- Graceful shutdown with signal handling
- Real-time statistics reporting
- Context-based cancellation

```bash
go run ./cmd/continuous_producer
```

### 2. Robust Producer (Enterprise Grade)
**Location**: `cmd/robust_producer/`
**Purpose**: Enterprise features with full observability
**Features**:
- Dead Letter Queue (DLQ)
- Prometheus metrics
- Error classification
- Health monitoring
- Circuit breaker patterns

```bash
go run ./cmd/robust_producer
```

### 3. Enhanced Producer (Development Demo)
**Location**: `cmd/enhanced_producer/`
**Purpose**: Demonstration of multiple production modes
**Features**:
- Hardcoded, random, and batch modes
- Multi-modal operation
- Development and testing focus

```bash
go run ./cmd/enhanced_producer
```

### 4. Basic Producer (Learning)
**Location**: `cmd/producer/`
**Purpose**: Simple producer for learning and basic testing
**Features**:
- Minimal implementation
- Clear, readable code
- Educational focus

```bash
go run ./cmd/producer
```

## 🔧 Development Workflow

### Adding New Features

1. **Libraries**: Add reusable code to `pkg/`
2. **Internal logic**: Add private code to `internal/`
3. **New applications**: Create in `cmd/new_app/`
4. **Tests**: Add to `test/` directory
5. **Documentation**: Update in `docs/`

### Testing Strategy

```bash
# Unit tests
go test ./pkg/...
go test ./internal/...

# Integration tests
go test ./test/...

# Specific producer tests
go test ./cmd/robust_producer/...
```

### Building and Deployment

```bash
# Build all producers
make build

# Build specific producer
go build -o bin/continuous_producer ./cmd/continuous_producer

# Docker deployment
docker-compose --profile continuous up
```

## 📚 Migration Notes

### For Existing Users

- **No breaking changes** to core functionality
- **All working code** is preserved in `cmd/`
- **Configuration remains the same** (`.env` files)
- **Docker setup unchanged**

### For Contributors

- **Follow Go standard layout**: https://github.com/golang-standards/project-layout
- **cmd/**: Executable applications
- **pkg/**: Library code that's safe for external use
- **internal/**: Private application and library code
- **test/**: Additional test files and test utilities

### Legacy Code Handling

Files in `legacy/` are:
- **Preserved for reference** but not actively maintained
- **Historical implementations** showing evolution
- **Safe to delete** if not needed
- **Not included in builds** or documentation

## 🔍 Key Benefits

### 1. **Clear Separation of Concerns**
- Applications vs libraries vs internal code
- Production vs development vs testing code
- Public vs private interfaces

### 2. **Better Developer Experience**
- Clear entry points in `cmd/`
- Reusable components in `pkg/`
- Organized test structure

### 3. **Improved Maintainability**
- No duplicate code
- Clear dependency boundaries
- Standard Go project layout

### 4. **Enhanced CI/CD**
- Separate builds for different applications
- Targeted testing strategies
- Clear deployment artifacts

## 🚀 Next Steps

1. **Complete protobuf generation** in devcontainer
2. **Update import paths** in applications (if needed)
3. **Add package documentation** to `pkg/` modules
4. **Create usage examples** in `examples/`
5. **Update CI/CD pipelines** for new structure

## 🤝 Contributing

When contributing to this project:

1. **Follow the new structure** for all additions
2. **Add tests** to the `test/` directory
3. **Document** new packages and applications
4. **Use the devcontainer** for consistent development environment

---

This reorganization transforms the repository from a "prototype with scattered files" to a "professional Go project with clear architecture." The structure now supports both learning and production use cases while maintaining all existing functionality.
