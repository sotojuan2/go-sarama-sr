# Enhanced Producer with Random Data Generation

This directory contains an enhanced version of the Kafka producer that integrates with the custom random shoe data generator.

## Features

The enhanced producer supports three different production modes:

### 1. Hardcoded Mode
- **Purpose**: Produces a single shoe with predefined values (original functionality)
- **Use Case**: Testing and validation with consistent data
- **Data**: Nike Air Max 90 - $89.99 - 4.5★

### 2. Random Mode  
- **Purpose**: Produces a single randomly generated shoe
- **Use Case**: Testing with realistic but varied data
- **Data**: Uses the custom generator following design specifications

### 3. Batch Mode
- **Purpose**: Produces multiple randomly generated shoes in sequence
- **Use Case**: Load testing and populating topics with diverse data
- **Data**: Configurable count (default: 3 shoes) with delays between messages

## Usage

### Run the Enhanced Producer
```bash
cd /workspaces/go-sarama-sr
go run cmd/enhanced_producer/main.go
```

The producer will automatically run all three modes in sequence, demonstrating the integration capabilities.

### Example Output
```
🔄 Running production mode: hardcoded
Hardcoded shoe: ID=12345, Brand=Nike, Name=Air Max 90, Price=89.99, Rating=4.5
✅ Message produced successfully to partition 3 at offset 0

🔄 Running production mode: random  
Random shoe: ID=1755527901661, Brand=Balenciaga, Name=Blazer Pro, Price=62.92, Rating=4.9
✅ Message produced successfully to partition 2 at offset 2

🔄 Running production mode: batch
Batch shoe 1: ID=1755527906958, Brand=Clarks, Name=Superstar 90, Price=323.96, Rating=4.2
Batch shoe 2: ID=1755527906909, Brand=Under Armour, Name=Boost QS Edition, Price=76.47, Rating=4.7
Batch shoe 3: ID=1755527906494, Brand=Vans, Name=Zoom Essential, Price=77.62, Rating=3.9
```

## Integration Features

- **Schema Registry**: All messages use the same `js_shoe` subject for schema registration
- **Error Handling**: Comprehensive error handling for serialization and production failures  
- **Logging**: Detailed logging for troubleshooting and monitoring
- **Flexibility**: Easy to modify batch sizes and add new production modes
- **Validation**: All generated data conforms to Protobuf schema requirements

## Data Quality

The random data generator ensures:
- **Unique IDs**: Timestamp-based with random suffixes
- **Realistic Brands**: Curated list of 20 authentic shoe brands
- **Pattern-based Names**: Multiple naming patterns for variety
- **Weighted Pricing**: Realistic distribution across price tiers
- **Skewed Ratings**: Distribution favoring higher ratings (like real products)

## Dependencies

- Uses the custom `pkg/generator` package for data generation
- Integrates with existing Schema Registry and Kafka infrastructure
- No external faker libraries required (pure Go standard library)
