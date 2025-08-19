#!/bin/bash

# Test script for Schema Registry Consumer
# Usage: ./test_consumer.sh

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

print_magic() {
    echo -e "${PURPLE}[MAGIC]${NC} $1"
}

print_data() {
    echo -e "${CYAN}[DATA]${NC} $1"
}

print_header() {
    echo -e "${BLUE}=== $1 ===${NC}"
}

# Function to parse Schema Registry data from console consumer output
parse_schema_registry_data() {
    while IFS=$'\t' read -r offset_part value_part; do
        # Extract offset number
        if [[ $offset_part =~ Offset:([0-9]+) ]]; then
            offset="${BASH_REMATCH[1]}"
        else
            continue
        fi
        
        # Convert escape sequences to hex
        hex_value=""
        i=0
        while [ $i -lt ${#value_part} ]; do
            char="${value_part:$i:1}"
            if [[ "$char" == "\\" && "${value_part:$((i+1)):1}" == "x" ]]; then
                # Extract hex value
                hex_char="${value_part:$((i+2)):2}"
                hex_value+="$hex_char"
                i=$((i+4))
            else
                # Convert regular character to hex
                printf -v hex_char "%02x" "'$char"
                hex_value+="$hex_char"
                i=$((i+1))
            fi
        done
        
        # Parse Schema Registry format
        if [[ ${#hex_value} -gt 10 ]]; then
            # Magic byte (first byte)
            magic_byte="${hex_value:0:2}"
            
            # Schema ID (next 4 bytes, big-endian)
            schema_id_hex="${hex_value:2:8}"
            
            # Convert hex to decimal for schema ID
            if [[ $schema_id_hex =~ ^[0-9A-Fa-f]{8}$ ]]; then
                schema_id=$((16#$schema_id_hex))
            else
                schema_id="INVALID"
            fi
            
            # Payload data (rest of the string)
            payload="${hex_value:10}"
            
            # Print formatted output
            echo -e "${CYAN}[OFFSET $offset]${NC} ${PURPLE}Magic: 0x$magic_byte${NC} ${YELLOW}Schema ID: $schema_id${NC}"
            
            # Additional analysis
            if [[ "$magic_byte" == "00" ]]; then
                print_magic "✅ Valid Schema Registry magic byte detected"
            else
                echo -e "${YELLOW}[WARN]${NC} Unexpected magic byte: 0x$magic_byte (expected: 0x00)"
            fi
            
            # Show payload length
            payload_length=$((${#payload} / 2))
            print_data "📏 Payload size: $payload_length bytes"
            
            # Show first few bytes of payload
            if [[ ${#payload} -gt 20 ]]; then
                print_data "👀 Payload preview: ${payload:0:20}..."
            else
                print_data "👀 Payload: $payload"
            fi
            
            # Try to extract readable brand/name if visible
            if [[ "$magic_byte" == "00" ]]; then
                # Look for readable strings in the payload (simple heuristic)
                readable=$(echo "$value_part" | grep -o '[A-Za-z][A-Za-z ]*[A-Za-z]' | head -2)
                if [[ -n "$readable" ]]; then
                    print_data "👟 Detected strings: $readable"
                fi
            fi
            
            echo -e "${BLUE}═══════════════════════════════════════${NC}"
        else
            echo -e "${YELLOW}[WARN]${NC} Message too short for Schema Registry format"
        fi
    done
}

print_header "Schema Registry Consumer Test"
print_status "Reading recent messages from js_shoe topic..."

# Load environment
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
ENV_FILE="$PROJECT_ROOT/.env"

if [[ -f "$ENV_FILE" ]]; then
    set -a
    source "$ENV_FILE"
    set +a
fi

# Run the consumer and parse output
kafka-console-consumer \
  --bootstrap-server "$KAFKA_BOOTSTRAP_SERVERS" \
  --topic js_shoe \
  --partition 0 \
  --offset earliest \
  --consumer.config cloud_sasl.config \
  --max-messages 3 \
  --timeout-ms 8000 \
  --property value.deserializer=org.apache.kafka.common.serialization.BytesDeserializer \
  --property print.value=true \
  --property print.offset=true 2>/dev/null | parse_schema_registry_data

print_status "✅ Schema Registry analysis completed!"
