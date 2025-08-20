package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/IBM/sarama"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde/protobuf"
	"github.com/go-sarama-sr/producer/internal/config"
	pb "github.com/go-sarama-sr/producer/pb"
	"github.com/go-sarama-sr/producer/pkg/generator"
	srClient "github.com/go-sarama-sr/producer/pkg/schemaregistry"
	"github.com/joho/godotenv"
)

// productionMode determines the data generation strategy
type productionMode string

const (
	HardcodedMode productionMode = "hardcoded"
	RandomMode    productionMode = "random"
	BatchMode     productionMode = "batch"
)

func main() {
	log.Println("Starting Enhanced Protobuf Message Producer...")

	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Could not load .env file: %v", err)
	}

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Create Schema Registry client
	srClientWrapper, err := srClient.NewClient(cfg)
	if err != nil {
		log.Fatalf("Error creating Schema Registry client: %v", err)
	}

	// Verify Schema Registry connectivity
	log.Println("Testing Schema Registry connectivity...")
	healthResult, err := srClientWrapper.ComprehensiveHealthCheck()
	if err != nil {
		log.Fatalf("Schema Registry health check failed: %v", err)
	}
	log.Printf("Schema Registry connectivity verified! Status: %+v", healthResult)

	// Initialize the random shoe generator
	shoeGen := generator.NewShoeGenerator()
	log.Println("Random shoe generator initialized")

	// Demo different production modes
	modes := []productionMode{HardcodedMode, RandomMode, BatchMode}

	for _, mode := range modes {
		log.Printf("\n🔄 Running production mode: %s", mode)

		switch mode {
		case HardcodedMode:
			err = runHardcodedMode(srClientWrapper, cfg)
		case RandomMode:
			err = runRandomMode(srClientWrapper, cfg, shoeGen)
		case BatchMode:
			err = runBatchMode(srClientWrapper, cfg, shoeGen, 3) // Generate 3 shoes
		}

		if err != nil {
			log.Printf("❌ Error in %s mode: %v", mode, err)
		} else {
			log.Printf("✅ %s mode completed successfully", mode)
		}

		// Small delay between modes
		time.Sleep(2 * time.Second)
	}

	log.Println("\n🎉 Enhanced Producer completed successfully!")
	fmt.Println("✅ Integration: Random data generation integrated with Kafka producer")
	fmt.Println("✅ Flexibility: Multiple production modes available")
	fmt.Println("✅ Validation: All modes tested and working")
}

// runHardcodedMode produces a single hardcoded shoe (original functionality)
func runHardcodedMode(client *srClient.Client, cfg *config.Config) error {
	log.Println("Creating hardcoded shoe...")

	shoe := &pb.Shoe{
		Id:        12345,
		Brand:     "Nike",
		Name:      "Air Max 90",
		SalePrice: 89.99,
		Rating:    4.5,
	}

	log.Printf("Hardcoded shoe: ID=%d, Brand=%s, Name=%s, Price=%.2f, Rating=%.1f",
		shoe.Id, shoe.Brand, shoe.Name, shoe.SalePrice, shoe.Rating)

	return produceShoeToKafka(client, cfg, shoe)
}

// runRandomMode produces a single randomly generated shoe
func runRandomMode(client *srClient.Client, cfg *config.Config, gen *generator.ShoeGenerator) error {
	log.Println("Generating random shoe...")

	shoe := gen.GenerateRandomShoe()

	log.Printf("Random shoe: ID=%d, Brand=%s, Name=%s, Price=%.2f, Rating=%.1f",
		shoe.Id, shoe.Brand, shoe.Name, shoe.SalePrice, shoe.Rating)

	return produceShoeToKafka(client, cfg, shoe)
}

// runBatchMode produces multiple randomly generated shoes
func runBatchMode(client *srClient.Client, cfg *config.Config, gen *generator.ShoeGenerator, count int) error {
	log.Printf("Generating batch of %d random shoes...", count)

	shoes := gen.GenerateRandomShoes(count)

	for i, shoe := range shoes {
		log.Printf("Batch shoe %d: ID=%d, Brand=%s, Name=%s, Price=%.2f, Rating=%.1f",
			i+1, shoe.Id, shoe.Brand, shoe.Name, shoe.SalePrice, shoe.Rating)

		err := produceShoeToKafka(client, cfg, shoe)
		if err != nil {
			return fmt.Errorf("failed to produce shoe %d: %w", i+1, err)
		}

		// Small delay between messages to avoid overwhelming Kafka
		time.Sleep(500 * time.Millisecond)
	}

	return nil
}

// produceShoeToKafka handles the complete process of serialization and production for a single shoe
func produceShoeToKafka(client *srClient.Client, cfg *config.Config, shoe *pb.Shoe) error {
	// Serialize with Schema Registry
	log.Println("Serializing shoe with Schema Registry...")
	serializedData, err := serializeShoeWithSchemaRegistry(client, shoe)
	if err != nil {
		return fmt.Errorf("serialization failed: %w", err)
	}
	log.Printf("Successfully serialized shoe data: %d bytes", len(serializedData))

	// Produce to Kafka
	log.Println("Producing message to Kafka...")
	err = produceMessageToKafka(cfg, serializedData)
	if err != nil {
		return fmt.Errorf("kafka production failed: %w", err)
	}

	return nil
}

// serializeShoeWithSchemaRegistry serializes the Shoe object using Schema Registry
func serializeShoeWithSchemaRegistry(client *srClient.Client, shoe *pb.Shoe) ([]byte, error) {
	log.Println("Creating Protobuf serializer...")

	// Configure Protobuf serializer
	serializer, err := protobuf.NewSerializer(client.GetClient(), serde.ValueSerde, protobuf.NewSerializerConfig())
	if err != nil {
		return nil, fmt.Errorf("failed to create protobuf serializer: %w", err)
	}

	// Serialize the object (schema is auto-registered if it doesn't exist)
	log.Println("Serializing shoe object (schema will be auto-registered)...")
	serializedData, err := serializer.Serialize("js_shoe", shoe)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize shoe: %w", err)
	}

	log.Printf("Schema registration and serialization completed successfully")
	return serializedData, nil
}

// produceMessageToKafka produces the serialized message to Kafka using Sarama
func produceMessageToKafka(cfg *config.Config, data []byte) error {
	log.Println("Setting up Kafka producer...")

	// Configure Sarama
	saramaConfig := sarama.NewConfig()
	saramaConfig.Producer.RequiredAcks = sarama.WaitForAll
	saramaConfig.Producer.Retry.Max = 5
	saramaConfig.Producer.Return.Successes = true
	saramaConfig.Producer.Return.Errors = true

	// SASL configuration for Confluent Cloud
	saramaConfig.Net.SASL.Enable = true
	saramaConfig.Net.SASL.Mechanism = sarama.SASLTypePlaintext
	saramaConfig.Net.SASL.User = cfg.Kafka.APIKey
	saramaConfig.Net.SASL.Password = cfg.Kafka.APISecret
	saramaConfig.Net.TLS.Enable = true

	// Create async producer
	brokers := strings.Split(cfg.Kafka.BootstrapServers, ",")
	producer, err := sarama.NewAsyncProducer(brokers, saramaConfig)
	if err != nil {
		return fmt.Errorf("failed to create async producer: %w", err)
	}
	defer producer.Close()

	// Create the message
	message := &sarama.ProducerMessage{
		Topic: cfg.Kafka.Topic,
		Value: sarama.ByteEncoder(data),
		Headers: []sarama.RecordHeader{
			{
				Key:   []byte("content-type"),
				Value: []byte("application/x-protobuf"),
			},
		},
	}

	log.Printf("Sending message to topic: %s", cfg.Kafka.Topic)

	// Send message
	producer.Input() <- message

	// Handle response
	select {
	case success := <-producer.Successes():
		log.Printf("✅ Message produced successfully to partition %d at offset %d",
			success.Partition, success.Offset)
		return nil
	case kafkaError := <-producer.Errors():
		return fmt.Errorf("failed to produce message: %w", kafkaError.Err)
	case <-time.After(30 * time.Second):
		return fmt.Errorf("timeout waiting for message production")
	}
}
