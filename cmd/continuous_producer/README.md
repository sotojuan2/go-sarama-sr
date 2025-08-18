# Continuous Kafka Producer

This directory contains the continuous Kafka producer that implements the final MVP functionality - an asynchronous loop that continuously generates random shoe data and produces it to Kafka at configurable intervals.

## Features

- **Continuous Operation**: Runs indefinitely until stopped
- **Asynchronous Production**: Non-blocking message production with proper error handling
- **Random Data Generation**: Uses the custom shoe generator for realistic, unique data
- **Schema Registry Integration**: Automatic schema registration and serialization
- **Graceful Shutdown**: Responds to SIGINT/SIGTERM with clean resource cleanup
- **Real-time Statistics**: Live metrics reporting including production rate and error counts
- **Configurable Timing**: Message production interval configurable via environment variables

## Configuration

Set these environment variables in your `.env` file:

```bash
# Message production interval (how often to send messages)
MESSAGE_INTERVAL=1s          # 1 second between messages (default)
# MESSAGE_INTERVAL=500ms     # 500 milliseconds for faster production
# MESSAGE_INTERVAL=5s        # 5 seconds for slower production

# Log level for verbosity control
LOG_LEVEL=info               # info (default), debug, warn, error

# Kafka and Schema Registry settings (same as other producers)
KAFKA_BOOTSTRAP_SERVERS=your-cluster.amazonaws.com:9092
KAFKA_API_KEY=your-api-key
KAFKA_API_SECRET=your-api-secret
KAFKA_TOPIC=js_shoe
SCHEMA_REGISTRY_URL=https://your-schema-registry.amazonaws.com
SCHEMA_REGISTRY_API_KEY=your-sr-api-key
SCHEMA_REGISTRY_API_SECRET=your-sr-api-secret
```

## Usage

### Build and Run

```bash
# Build the continuous producer
go build -o continuous_producer ./cmd/continuous_producer

# Run the continuous producer
./continuous_producer
```

### Direct Execution

```bash
# Run directly with Go
go run ./cmd/continuous_producer/main.go
```

### Sample Output

```
🚀 Starting Continuous Kafka Producer with Random Shoe Data Generation...
📊 Configuration loaded:
   📡 Kafka Cluster: pkc-xxxxx.europe-southwest1.gcp.confluent.cloud:9092
   📝 Topic: js_shoe
   ⏱️ Message Interval: 1s
   📈 Schema Registry: https://psrc-xxxxx.eu-west-2.aws.confluent.cloud
🔧 Initializing continuous producer components...
🔍 Testing Schema Registry connectivity...
✅ Schema Registry connectivity verified! Status: {Status:healthy}
✅ Random shoe generator initialized
✅ Protobuf serializer created
✅ Kafka async producer created
🔄 Starting continuous production loop...
   Press Ctrl+C to stop gracefully
🔄 Production loop started with interval: 1s
📊 Stats: Produced=10, Errors=0, Rate=1.00 msgs/sec, Runtime=10s
📊 Stats: Produced=20, Errors=0, Rate=1.00 msgs/sec, Runtime=20s
^C
⏹️ Shutdown signal received. Stopping production...
🛑 Production loop stopping...
✅ Graceful shutdown completed
📊 Final Statistics:
   📤 Messages Produced: 25
   ❌ Errors: 0
   ⏱️ Runtime: 25s
   📈 Average Rate: 1.00 messages/second
```

## Architecture

### Components

1. **ContinuousProducer**: Main coordinator managing all components
2. **Production Loop**: Timer-based loop generating and sending messages
3. **Success Handler**: Goroutine processing successful message deliveries
4. **Error Handler**: Goroutine processing message delivery errors
5. **Stats Reporter**: Goroutine providing real-time production metrics
6. **Graceful Shutdown**: Signal handling for clean termination

### Message Flow

1. **Timer Trigger**: Every `MESSAGE_INTERVAL`, generate new shoe data
2. **Data Generation**: Create random shoe using custom generator
3. **Serialization**: Serialize shoe with Schema Registry (auto-registration)
4. **Async Production**: Queue message to Kafka asynchronously
5. **Result Handling**: Process success/error responses in separate goroutines
6. **Statistics**: Track and report production metrics

### Error Handling

- **Serialization Errors**: Logged and counted, loop continues
- **Kafka Errors**: Processed by dedicated error handler goroutine
- **Network Issues**: Automatic retries via Sarama configuration
- **Graceful Shutdown**: Clean resource cleanup on termination signals

## Integration

The continuous producer integrates:

- **Custom Generator** (`pkg/generator`): For realistic random shoe data
- **Schema Registry Client** (`pkg/schemaregistry`): For schema management
- **Configuration System** (`internal/config`): For environment-based setup
- **Protobuf Schemas** (`pb/`): For message structure definition

## Monitoring

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

## Performance

- **High Throughput**: Asynchronous production with minimal blocking
- **Low Latency**: Efficient serialization and message queuing
- **Memory Efficient**: Proper resource cleanup and goroutine management
- **Scalable**: Configurable production rate and batch handling

## Task Completion

This producer completes **Task 7: Integrate Random Data Generation into Asynchronous Producer Loop**:

✅ **7.1**: Asynchronous producer loop implemented  
✅ **7.2**: Random shoe data generation integrated  
✅ **7.3**: Messages serialized and produced asynchronously  
✅ **7.4**: Loop timing and message frequency managed  
✅ **7.5**: Kafka topic monitoring enabled through logs and stats  

The MVP is now complete with a fully functional continuous data producer.
