#!/bin/bash

# Script to generate cloud_sasl.config from .env file
# Usage: ./generate_config.sh

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_header() {
    echo -e "${BLUE}=== $1 ===${NC}"
}

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
ENV_FILE="$PROJECT_ROOT/.env"
CONFIG_FILE="$SCRIPT_DIR/cloud_sasl.config"

print_header "Confluent Cloud SASL Configuration Generator"

# Check if .env file exists
if [[ ! -f "$ENV_FILE" ]]; then
    print_error ".env file not found at: $ENV_FILE"
    print_error "Please ensure you have a .env file with KAFKA_API_KEY and KAFKA_API_SECRET"
    exit 1
fi

print_status "Found .env file at: $ENV_FILE"

# Source the .env file
print_status "Loading environment variables..."
set -a  # Export all variables
source "$ENV_FILE"
set +a  # Stop exporting

# Validate required variables
if [[ -z "$KAFKA_API_KEY" ]]; then
    print_error "KAFKA_API_KEY not found in .env file"
    exit 1
fi

if [[ -z "$KAFKA_API_SECRET" ]]; then
    print_error "KAFKA_API_SECRET not found in .env file"
    exit 1
fi

# Optional: Check for bootstrap servers
if [[ -z "$KAFKA_BOOTSTRAP_SERVERS" ]]; then
    print_warning "KAFKA_BOOTSTRAP_SERVERS not found in .env file"
    print_warning "You'll need to specify bootstrap servers manually"
fi

print_status "Validated required environment variables"
print_status "  - KAFKA_API_KEY: ${KAFKA_API_KEY:0:8}..."
print_status "  - KAFKA_API_SECRET: [HIDDEN]"

# Generate cloud_sasl.config
print_status "Generating cloud_sasl.config..."

cat > "$CONFIG_FILE" << EOF
# Kafka SASL Configuration for Confluent Cloud
# Generated on: $(date)
# From .env file: $ENV_FILE

# SASL Configuration
security.protocol=SASL_SSL
sasl.mechanism=PLAIN
sasl.jaas.config=org.apache.kafka.common.security.plain.PlainLoginModule required username="$KAFKA_API_KEY" password="$KAFKA_API_SECRET";

# SSL Configuration  
ssl.endpoint.identification.algorithm=https

# Consumer specific settings (optional)
auto.offset.reset=earliest
enable.auto.commit=true
# group.id=console-consumer-dynamic  # Will be overridden by --group parameter
EOF

print_status "Generated configuration file: $CONFIG_FILE"

# Validate the generated file
if [[ -f "$CONFIG_FILE" ]]; then
    print_status "Configuration file created successfully!"
    echo ""
    print_header "Generated Configuration Preview"
    echo "File: $CONFIG_FILE"
    echo ""
    # Show config but hide the credentials
    sed 's/password="[^"]*"/password="[HIDDEN]"/' "$CONFIG_FILE"
    echo ""
    
    print_header "Usage Examples"
    echo ""
    echo "1. Basic console consumer:"
    echo "   kafka-console-consumer \\"
    echo "     --bootstrap-server \$KAFKA_BOOTSTRAP_SERVERS \\"
    echo "     --topic \$KAFKA_TOPIC \\"
    echo "     --consumer.config $CONFIG_FILE \\"
    echo "     --from-beginning"
    echo ""
    
    echo "2. Schema Registry analysis (use the provided script):"
    echo "   ./schema_registry_consumer.sh"
    echo ""
    
    print_status "Configuration ready for use!"
else
    print_error "Failed to create configuration file"
    exit 1
fi
