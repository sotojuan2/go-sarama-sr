#!/bin/bash

# Schema Registry Consumer with Magic Byte Analysis
# Combines Kafka console consumer with Schema Registry magic byte parsing
# Usage: ./schema_registry_consumer.sh [topic] [--from-beginning|--from-latest]

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
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

print_magic() {
    echo -e "${PURPLE}[MAGIC]${NC} $1"
}

print_data() {
    echo -e "${CYAN}[DATA]${NC} $1"
}

# Default values
DEFAULT_TOPIC="js_shoe"
DEFAULT_OFFSET_MODE="--from-beginning"

# Parse command line arguments
TOPIC="${1:-$DEFAULT_TOPIC}"
OFFSET_MODE="${2:-$DEFAULT_OFFSET_MODE}"

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
ENV_FILE="$PROJECT_ROOT/.env"
CONFIG_FILE="$SCRIPT_DIR/cloud_sasl.config"

print_header "Schema Registry Consumer with Magic Byte Analysis"

# Check if .env file exists and load it
if [[ ! -f "$ENV_FILE" ]]; then
    print_error ".env file not found at: $ENV_FILE"
    exit 1
fi

# Load environment variables
print_status "Loading environment variables from .env..."
set -a
source "$ENV_FILE"
set +a

# Validate required variables
if [[ -z "$KAFKA_BOOTSTRAP_SERVERS" ]]; then
    print_error "KAFKA_BOOTSTRAP_SERVERS not found in .env file"
    exit 1
fi

if [[ -z "$KAFKA_API_KEY" ]] || [[ -z "$KAFKA_API_SECRET" ]]; then
    print_error "KAFKA_API_KEY and KAFKA_API_SECRET required in .env file"
    exit 1
fi

print_status "Configuration loaded:"
print_status "  - Bootstrap Servers: $KAFKA_BOOTSTRAP_SERVERS"
print_status "  - Topic: $TOPIC"
print_status "  - Offset Mode: $OFFSET_MODE"

# Generate config if it doesn't exist
if [[ ! -f "$CONFIG_FILE" ]]; then
    print_status "Config file not found, generating..."
    "$SCRIPT_DIR/generate_config.sh"
fi

print_status "Using config file: $CONFIG_FILE"

# Function to parse magic bytes and schema ID
parse_schema_registry_data() {
    while IFS= read -r line; do
        # Extract offset and value from the line
        if [[ $line =~ ^([0-9]+)[[:space:]]+(.+)$ ]]; then
            offset="${BASH_REMATCH[1]}"
            value="${BASH_REMATCH[2]}"
            
            # Parse Schema Registry format
            if [[ ${#value} -gt 10 ]]; then
                # Magic byte (first 2 hex chars = 1 byte)
                magic_byte="${value:0:2}"
                
                # Schema ID (next 8 hex chars = 4 bytes, big-endian)
                schema_id_hex="${value:2:8}"
                
                # Convert hex to decimal for schema ID
                if [[ $schema_id_hex =~ ^[0-9A-Fa-f]{8}$ ]]; then
                    schema_id=$((16#$schema_id_hex))
                else
                    schema_id="INVALID"
                fi
                
                # Payload data (rest of the string)
                payload="${value:10}"
                
                # Print formatted output
                echo -e "${CYAN}[OFFSET $offset]${NC} ${PURPLE}Magic: 0x$magic_byte${NC} ${YELLOW}Schema ID: $schema_id${NC}"
                
                # Additional analysis
                if [[ "$magic_byte" == "00" ]]; then
                    print_magic "Valid Schema Registry magic byte detected"
                else
                    print_warning "Unexpected magic byte: 0x$magic_byte (expected: 0x00)"
                fi
                
                # Show payload length
                payload_length=$((${#payload} / 2))
                print_data "Payload size: $payload_length bytes"
                
                # Show first few bytes of payload
                if [[ ${#payload} -gt 20 ]]; then
                    print_data "Payload preview: ${payload:0:20}..."
                else
                    print_data "Payload: $payload"
                fi
                
                echo "---"
            else
                print_warning "Message too short for Schema Registry format: $value"
            fi
        else
            print_warning "Unable to parse line: $line"
        fi
    done
}

# Function to show help
show_help() {
    echo "Usage: $0 [topic] [offset_mode]"
    echo ""
    echo "Arguments:"
    echo "  topic        Topic name (default: $DEFAULT_TOPIC)"
    echo "  offset_mode  --from-beginning or --from-latest (default: $DEFAULT_OFFSET_MODE)"
    echo ""
    echo "Examples:"
    echo "  $0                                    # Use defaults"
    echo "  $0 my-topic                          # Custom topic"
    echo "  $0 my-topic --from-latest           # Custom topic, latest offset"
    echo ""
    echo "Environment variables required in .env:"
    echo "  KAFKA_BOOTSTRAP_SERVERS"
    echo "  KAFKA_API_KEY"
    echo "  KAFKA_API_SECRET"
}

# Check for help flag
if [[ "$1" == "-h" ]] || [[ "$1" == "--help" ]]; then
    show_help
    exit 0
fi

# Validate offset mode
if [[ "$OFFSET_MODE" != "--from-beginning" ]] && [[ "$OFFSET_MODE" != "--from-latest" ]]; then
    print_error "Invalid offset mode: $OFFSET_MODE"
    print_error "Use --from-beginning or --from-latest"
    exit 1
fi

print_header "Starting Schema Registry Consumer"
print_status "Press Ctrl+C to stop the consumer"
echo ""

# Check if kafka-console-consumer is available
if ! command -v kafka-console-consumer &> /dev/null; then
    print_error "kafka-console-consumer not found in PATH"
    print_error "Please ensure Kafka CLI tools are installed and available"
    exit 1
fi

# Build the consumer command
CONSUMER_CMD="kafka-console-consumer \
    --bootstrap-server $KAFKA_BOOTSTRAP_SERVERS \
    --topic $TOPIC \
    --consumer.config $CONFIG_FILE \
    $OFFSET_MODE \
    --property value.deserializer=org.apache.kafka.common.serialization.BytesDeserializer \
    --property print.value=true \
    --property print.offset=true"

print_status "Executing consumer command..."
print_status "Command: $CONSUMER_CMD"
echo ""

# Execute the consumer with magic byte parsing
eval "$CONSUMER_CMD" | parse_schema_registry_data

print_status "Consumer stopped."
