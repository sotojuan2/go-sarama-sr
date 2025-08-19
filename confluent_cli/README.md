# Confluent CLI Tools

This directory contains scripts and tools for consuming and analyzing Kafka messages from Confluent Cloud with Schema Registry integration.

## 🎯 Overview

These tools automate the process of:
1. **Reading credentials** from your `.env` file
2. **Generating SASL configuration** for Confluent Cloud
3. **Consuming messages** with Schema Registry magic byte analysis
4. **Parsing Schema Registry format** (magic bytes + schema ID + payload)
5. **Verifying Schema Registry integration** in your events

## ✅ **Schema Registry Verification**

This tool is **the definitive way to verify** that your Kafka events are properly using Schema Registry integration. When events are serialized with Schema Registry, they contain:

- **Magic Byte**: `0x00` (indicates Schema Registry format)
- **Schema ID**: 4-byte schema identifier (e.g., `100190`)
- **Payload**: Actual serialized data (Protobuf/Avro/JSON Schema)

**Why this verification matters:**
- ✅ Confirms your producers are using Schema Registry correctly
- ✅ Validates schema evolution compatibility
- ✅ Proves data governance and schema enforcement
- ✅ Essential for Client-Side Field-Level Encryption (CSFLE) setups

## 📁 Files

| File | Purpose | Git Tracked |
|------|---------|-------------|
| `generate_config.sh` | Generates `cloud_sasl.config` from `.env` variables | ✅ Yes |
| `schema_registry_consumer.sh` | Main consumer with Schema Registry analysis | ✅ Yes |
| `cloud_sasl.config.example` | Example configuration file (safe template) | ✅ Yes |
| `cloud_sasl.config` | **Generated SASL configuration (contains real credentials)** | 🚫 **NO - Protected by .gitignore** |
| `README.md` | This documentation | ✅ Yes |

## 🔐 **Security Notes**

⚠️  **IMPORTANT**: The `cloud_sasl.config` file contains your real API credentials and is automatically excluded from Git via `.gitignore`. This file:

- ✅ **Is generated automatically** by `generate_config.sh`
- ✅ **Stays local** and never gets committed to GitHub
- ✅ **Contains your real credentials** loaded from `.env`
- ✅ **Is safe to regenerate** anytime

**Files that ARE safe to commit:**
- All `.sh` scripts (no credentials hardcoded)
- `cloud_sasl.config.example` (template with placeholder values)
- `README.md` and documentation

**Files that are PROTECTED (never committed):**
- `cloud_sasl.config` (contains real credentials)
- `.env` file (contains API keys)
- Any `*.config` files with real credentials

## 🚀 Quick Start

### 1. Prerequisites

Ensure your `.env` file contains:
```bash
KAFKA_BOOTSTRAP_SERVERS=pkc-xxxxx.region.provider.confluent.cloud:9092
KAFKA_API_KEY=your-kafka-api-key
KAFKA_API_SECRET=your-kafka-api-secret
KAFKA_TOPIC=js_shoe
```

### 2. Make Scripts Executable

```bash
chmod +x confluent_cli/*.sh
```

### 3. Generate Configuration

```bash
cd confluent_cli
./generate_config.sh
```

This creates `cloud_sasl.config` with your credentials.

### 4. Consume Messages with Schema Registry Analysis

```bash
# Use defaults (js_shoe topic, from beginning)
./schema_registry_consumer.sh

# Custom topic
./schema_registry_consumer.sh my-topic

# Custom topic, latest messages only
./schema_registry_consumer.sh my-topic --from-latest
```

## 🔍 Schema Registry Analysis

The consumer automatically parses Schema Registry wire format:

```
[OFFSET 123] Magic: 0x00 Schema ID: 100004
[MAGIC] Valid Schema Registry magic byte detected
[DATA] Payload size: 142 bytes
[DATA] Payload preview: 0a4e696b651a0d4169...
---
```

### Message Format Analysis

- **Magic Byte**: `0x00` (indicates Schema Registry format)
- **Schema ID**: 4-byte big-endian integer (schema version)
- **Payload**: Actual Protobuf-encoded message data

## 🎉 **Verified Working Results**

### ✅ **Real Test Results (js_shoe topic)**

Successful verification of Schema Registry integration:

```bash
=== Schema Registry Consumer Test ===
[INFO] Reading recent messages from js_shoe topic...

[OFFSET 0] Magic: 0x00 Schema ID: 100190
[MAGIC] ✅ Valid Schema Registry magic byte detected
[DATA] 📏 Payload size: 47 bytes
[DATA] 👀 Payload preview: 000895B4E0D9D096BBAE...
═══════════════════════════════════════

[OFFSET 1] Magic: 0x00 Schema ID: 100190
[MAGIC] ✅ Valid Schema Registry magic byte detected
[DATA] 📏 Payload size: 54 bytes
[DATA] 👀 Payload preview: 000884FFD1B7D496BBAE...
═══════════════════════════════════════

[OFFSET 2] Magic: 0x00 Schema ID: 100190
[MAGIC] ✅ Valid Schema Registry magic byte detected
[DATA] 📏 Payload size: 47 bytes
[DATA] 👀 Payload preview: 0008F1CBBAD5D696BBAE...
═══════════════════════════════════════

[INFO] ✅ Schema Registry analysis completed!
```

**🎯 Analysis Results:**
- ✅ **All events are using Schema Registry**: Magic byte `0x00` detected in all messages
- ✅ **Consistent schema version**: Schema ID `100190` across all messages
- ✅ **Protobuf serialization**: Variable payload sizes (47-54 bytes) indicate proper protobuf encoding
- ✅ **Data governance active**: Schema enforcement is working correctly

### 🔍 **How to Interpret Results**

#### ✅ **Schema Registry Confirmed**
```bash
[OFFSET X] Magic: 0x00 Schema ID: 100190
[MAGIC] ✅ Valid Schema Registry magic byte detected
```
**Meaning**: Event is properly serialized with Schema Registry

#### ❌ **Non-Schema Registry Event**
```bash
[OFFSET X] Magic: 0x7b Schema ID: 22636f6d
[WARN] Unexpected magic byte: 0x7b (expected: 0x00)
```
**Meaning**: Event was NOT serialized with Schema Registry (plain JSON/text)

### 🎯 **Use Cases for Verification**

1. **Protobuf Migration**: Verify all producers switched to protobuf+SR
2. **Schema Evolution**: Confirm schema versions are consistent
3. **Data Governance**: Validate schema enforcement policies
4. **CSFLE Setup**: Prerequisite for Client-Side Field-Level Encryption
5. **Debugging**: Identify producers not using Schema Registry

## 📊 Additional Output Examples

## 🛠️ Manual Usage

If you prefer manual control:

### 1. Generate Config Only
```bash
./generate_config.sh
```

### 2. Use Standard Kafka Tools
```bash
# Basic consumer
kafka-console-consumer \
  --bootstrap-server $KAFKA_BOOTSTRAP_SERVERS \
  --topic $KAFKA_TOPIC \
  --consumer.config cloud_sasl.config \
  --from-beginning

# Consumer with byte deserializer
kafka-console-consumer \
  --bootstrap-server $KAFKA_BOOTSTRAP_SERVERS \
  --topic $KAFKA_TOPIC \
  --consumer.config cloud_sasl.config \
  --from-beginning \
  --property value.deserializer=org.apache.kafka.common.serialization.BytesDeserializer \
  --property print.value=true \
  --property print.offset=true
```

## 🔧 Configuration Details

### Generated `cloud_sasl.config`
```properties
# Kafka SASL Configuration for Confluent Cloud
security.protocol=SASL_SSL
sasl.mechanism=PLAIN
sasl.jaas.config=org.apache.kafka.common.security.plain.PlainLoginModule required username="your-api-key" password="your-api-secret";

# SSL Configuration  
ssl.endpoint.identification.algorithm=https

# Consumer specific settings
auto.offset.reset=earliest
enable.auto.commit=true
group.id=schema-registry-console-consumer
```

### Environment Variables Used

| Variable | Required | Description |
|----------|----------|-------------|
| `KAFKA_BOOTSTRAP_SERVERS` | ✅ | Kafka cluster bootstrap servers |
| `KAFKA_API_KEY` | ✅ | Kafka API key for SASL authentication |
| `KAFKA_API_SECRET` | ✅ | Kafka API secret for SASL authentication |
| `KAFKA_TOPIC` | ❌ | Default topic (fallback: `js_shoe`) |

## 🎨 Features

### 🔐 Security
- **No hardcoded credentials**: All credentials from `.env`
- **Hidden secrets**: Passwords masked in output
- **Secure config generation**: Proper SASL_SSL configuration

### 🎯 Schema Registry Integration
- **Magic byte detection**: Validates Schema Registry format
- **Schema ID extraction**: Parses 4-byte schema identifier
- **Payload analysis**: Shows message size and preview
- **Error handling**: Graceful handling of malformed messages

### 🖥️ User Experience
- **Colored output**: Easy-to-read colored console output
- **Flexible usage**: Multiple consumption modes
- **Help system**: Built-in help with examples
- **Error validation**: Clear error messages and validation

### 🔄 Automation
- **Auto-config generation**: Creates SASL config if missing
- **Environment integration**: Seamless `.env` file integration
- **Makefile integration**: Can be integrated with project Makefile

## 🧪 Testing

### Test Configuration Generation
```bash
# Test config generation
./generate_config.sh

# Verify config file
cat cloud_sasl.config
```

### Test Consumer (Dry Run)
```bash
# Test with help
./schema_registry_consumer.sh --help

# Test configuration loading
./schema_registry_consumer.sh js_shoe --from-latest
```

## 🔗 Integration with Project

### Makefile Integration
Add to your main `Makefile`:

```makefile
# Consumer targets
.PHONY: consumer-config consumer-run consumer-latest

consumer-config: ## Generate Confluent Cloud consumer configuration
	@./confluent_cli/generate_config.sh

consumer-run: consumer-config ## Run Schema Registry consumer (from beginning)
	@./confluent_cli/schema_registry_consumer.sh

consumer-latest: consumer-config ## Run Schema Registry consumer (latest only)
	@./confluent_cli/schema_registry_consumer.sh $(KAFKA_TOPIC) --from-latest

consumer-topic: consumer-config ## Run consumer for specific topic (usage: make consumer-topic TOPIC=my-topic)
	@./confluent_cli/schema_registry_consumer.sh $(TOPIC) --from-beginning
```

### Usage Examples
```bash
# Generate config and consume
make consumer-run

# Consume latest messages only
make consumer-latest

# Consume specific topic
make consumer-topic TOPIC=my-custom-topic
```

## 🚨 Troubleshooting

### Common Issues

1. **Config file not found**
   ```bash
   [ERROR] .env file not found
   ```
   **Solution**: Ensure `.env` file exists in project root

2. **Missing credentials**
   ```bash
   [ERROR] KAFKA_API_KEY not found in .env file
   ```
   **Solution**: Add required variables to `.env`

3. **Kafka tools not found**
   ```bash
   [ERROR] kafka-console-consumer not found in PATH
   ```
   **Solution**: Install Confluent CLI or Kafka tools

4. **Connection issues**
   ```bash
   ERROR Timeout waiting for response
   ```
   **Solution**: Check bootstrap servers and credentials

### Debugging

Enable verbose output:
```bash
set -x  # Enable bash debugging
./schema_registry_consumer.sh
set +x  # Disable debugging
```

## 📚 References

- [Confluent Cloud CLI](https://docs.confluent.io/confluent-cli/current/overview.html)
- [Kafka Console Consumer](https://kafka.apache.org/documentation/#basic_ops_consumer)
- [Schema Registry Wire Format](https://docs.confluent.io/platform/current/schema-registry/serdes-develop/index.html#wire-format)
- [SASL/SCRAM Authentication](https://docs.confluent.io/cloud/current/access-management/authenticate/api-keys/api-keys.html)

---

**🎉 Ready to consume and analyze your Schema Registry messages!**
