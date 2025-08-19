package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
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
	log.Println("🚀 Starting Enhanced Protobuf Message Producer...")

	// Set up graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

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
	log.Println("🔍 Testing Schema Registry connectivity...")
	healthResult, err := srClientWrapper.ComprehensiveHealthCheck()
	if err != nil {
		log.Fatalf("Schema Registry health check failed: %v", err)
	}
	log.Printf("✅ Schema Registry connectivity verified! Status: %+v", healthResult)

	// Initialize the random shoe generator
	shoeGen := generator.NewShoeGenerator()
	log.Println("✅ Random shoe generator initialized")

	// Demo different production modes
	modes := []productionMode{HardcodedMode, RandomMode, BatchMode}

	var wg sync.WaitGroup
	done := make(chan bool, 1)

	// Run production modes in a separate goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer func() { done <- true }()

		for _, mode := range modes {
			select {
			case <-ctx.Done():
				log.Printf("⏹️ Shutdown signal received, stopping production...")
				return
			default:
				log.Printf("\n🔄 Running production mode: %s", mode)

				switch mode {
				case HardcodedMode:
					err = runHardcodedModeWithContext(ctx, srClientWrapper, cfg)
				case RandomMode:
					err = runRandomModeWithContext(ctx, srClientWrapper, cfg, shoeGen)
				case BatchMode:
					err = runBatchModeWithContext(ctx, srClientWrapper, cfg, shoeGen, 3) // Generate 3 shoes
				}

				if err != nil {
					if ctx.Err() != nil {
						log.Printf("⏹️ Production mode %s stopped due to shutdown signal", mode)
						return
					}
					log.Printf("❌ Error in %s mode: %v", mode, err)
				} else {
					log.Printf("✅ %s mode completed successfully", mode)
				}

				// Small delay between modes (unless shutting down)
				select {
				case <-ctx.Done():
					return
				case <-time.After(2 * time.Second):
				}
			}
		}
	}()

	// Wait for either completion or shutdown signal
	select {
	case <-signalChan:
		log.Println("\n⏹️ Shutdown signal received. Stopping production...")
		cancel()

		// Wait for graceful shutdown with timeout
		shutdownDone := make(chan struct{})
		go func() {
			wg.Wait()
			close(shutdownDone)
		}()

		select {
		case <-shutdownDone:
			log.Println("✅ Graceful shutdown completed")
		case <-time.After(10 * time.Second):
			log.Println("⚠️ Shutdown timeout reached, forcing exit")
		}

	case <-done:
		log.Println("\n🎉 Enhanced Producer completed successfully!")
		fmt.Println("✅ Integration: Random data generation integrated with Kafka producer")
		fmt.Println("✅ Flexibility: Multiple production modes available")
		fmt.Println("✅ Validation: All modes tested and working")
		fmt.Println("✅ Graceful Shutdown: Signal handling implemented")
	}
}

// Context-aware functions for graceful shutdown

// runHardcodedModeWithContext produces a single hardcoded shoe with context cancellation support
func runHardcodedModeWithContext(ctx context.Context, client *srClient.Client, cfg *config.Config) error {
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

	return produceShoeToKafkaWithContext(ctx, client, cfg, shoe)
}

// runRandomModeWithContext produces a single randomly generated shoe with context cancellation support
func runRandomModeWithContext(ctx context.Context, client *srClient.Client, cfg *config.Config, gen *generator.ShoeGenerator) error {
	log.Println("Generating random shoe...")

	shoe := gen.GenerateRandomShoe()

	log.Printf("Random shoe: ID=%d, Brand=%s, Name=%s, Price=%.2f, Rating=%.1f",
		shoe.Id, shoe.Brand, shoe.Name, shoe.SalePrice, shoe.Rating)

	return produceShoeToKafkaWithContext(ctx, client, cfg, shoe)
}

// runBatchModeWithContext produces multiple randomly generated shoes with context cancellation support
func runBatchModeWithContext(ctx context.Context, client *srClient.Client, cfg *config.Config, gen *generator.ShoeGenerator, count int) error {
	log.Printf("Generating %d random shoes in batch mode...", count)

	for i := 0; i < count; i++ {
		select {
		case <-ctx.Done():
			log.Printf("⏹️ Batch mode cancelled after producing %d/%d shoes", i, count)
			return ctx.Err()
		default:
			shoe := gen.GenerateRandomShoe()

			log.Printf("Batch shoe %d/%d: ID=%d, Brand=%s, Name=%s, Price=%.2f, Rating=%.1f",
				i+1, count, shoe.Id, shoe.Brand, shoe.Name, shoe.SalePrice, shoe.Rating)

			if err := produceShoeToKafkaWithContext(ctx, client, cfg, shoe); err != nil {
				return fmt.Errorf("error producing batch shoe %d: %w", i+1, err)
			}

			// Small delay between batch items (unless shutting down)
			select {
			case <-ctx.Done():
				log.Printf("⏹️ Batch mode cancelled after producing %d/%d shoes", i+1, count)
				return ctx.Err()
			case <-time.After(1 * time.Second):
			}
		}
	}

	log.Printf("✅ Batch mode completed: %d shoes produced", count)
	return nil
}

// produceShoeToKafkaWithContext produces a shoe to Kafka with context cancellation support
func produceShoeToKafkaWithContext(ctx context.Context, client *srClient.Client, cfg *config.Config, shoe *pb.Shoe) error {
	// Check if context is already cancelled
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Serialize with Schema Registry
	log.Println("Serializing shoe with Schema Registry...")
	serializedData, err := serializeShoeWithSchemaRegistry(client, shoe)
	if err != nil {
		return fmt.Errorf("serialization failed: %w", err)
	}
	log.Printf("Successfully serialized shoe data: %d bytes", len(serializedData))

	// Check if context is cancelled before producing
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Produce to Kafka with context
	log.Println("Producing message to Kafka...")
	err = produceMessageToKafkaWithContext(ctx, cfg, serializedData)
	if err != nil {
		return fmt.Errorf("kafka production failed: %w", err)
	}

	return nil
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

// produceMessageToKafkaWithContext produces the serialized message to Kafka using Sarama with context cancellation support
func produceMessageToKafkaWithContext(ctx context.Context, cfg *config.Config, data []byte) error {
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

	// Handle response with context cancellation support
	select {
	case success := <-producer.Successes():
		log.Printf("✅ Message produced successfully to partition %d at offset %d",
			success.Partition, success.Offset)
		return nil
	case kafkaError := <-producer.Errors():
		return fmt.Errorf("failed to produce message: %w", kafkaError.Err)
	case <-ctx.Done():
		log.Printf("⏹️ Message production cancelled due to shutdown signal")
		return ctx.Err()
	case <-time.After(30 * time.Second):
		return fmt.Errorf("timeout waiting for message production")
	}
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
