package main

import (
	"fmt"
	"log"
	"time"

	"github.com/IBM/sarama"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde/protobuf"
	pb "github.com/go-sarama-sr/producer/pb"
	"github.com/go-sarama-sr/producer/internal/config"
	srClient "github.com/go-sarama-sr/producer/pkg/schemaregistry"
)

// createHardcodedShoe creates a sample Shoe object with realistic data
func createHardcodedShoe() *pb.Shoe {
	shoe := &pb.Shoe{
		Id:        12345,
		Brand:     "Nike",
		Name:      "Air Max 90",
		SalePrice: 89.99,
		Rating:    4.5,
	}

	log.Printf("Created hardcoded shoe: ID=%d, Brand=%s, Name=%s, Price=%.2f, Rating=%.1f",
		shoe.Id, shoe.Brand, shoe.Name, shoe.SalePrice, shoe.Rating)

	return shoe
}

func main() {
	log.Println("Starting Protobuf Message Producer...")

	// Cargar configuración
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Crear cliente del Schema Registry
	srClientWrapper, err := srClient.NewClient(cfg)
	if err != nil {
		log.Fatalf("Error creating Schema Registry client: %v", err)
	}

	// Subtask 5.1: Create hardcoded Shoe object
	shoe := createHardcodedShoe()
	log.Printf("Successfully created shoe object: %+v", shoe)

	// Verificar conectividad con el Schema Registry
	log.Println("Testing Schema Registry connectivity...")
	healthResult, err := srClientWrapper.ComprehensiveHealthCheck()
	if err != nil {
		log.Fatalf("Schema Registry health check failed: %v", err)
	}
	log.Printf("Schema Registry connectivity verified! Status: %+v", healthResult)

	// Subtask 5.2 & 5.3: Serialize with Schema Registry (schema auto-registration)
	serializedData, err := serializeShoeWithSchemaRegistry(srClientWrapper, shoe)
	if err != nil {
		log.Fatalf("Error serializing shoe: %v", err)
	}
	log.Printf("Successfully serialized shoe data: %d bytes", len(serializedData))

	// Subtask 5.4: Produce message to Kafka
	err = produceMessageToKafka(cfg, serializedData)
	if err != nil {
		log.Fatalf("Error producing message to Kafka: %v", err)
	}

	// Subtask 5.5: Success confirmation
	log.Println("✅ All subtasks completed successfully!")
	fmt.Println("🎉 Message produced to Kafka topic successfully!")
	fmt.Println("📊 You can verify the message in Confluent Cloud UI or with a consumer")
}

// serializeShoeWithSchemaRegistry serializes the Shoe object using Schema Registry
func serializeShoeWithSchemaRegistry(client *srClient.Client, shoe *pb.Shoe) ([]byte, error) {
	log.Println("Serializing shoe with Schema Registry...")
	
	// Create Protobuf serializer
	serializer, err := protobuf.NewSerializer(client.GetClient(), serde.ValueSerde, protobuf.NewSerializerConfig())
	if err != nil {
		return nil, fmt.Errorf("failed to create protobuf serializer: %w", err)
	}

	// Serialize the object (schema is auto-registered if not exists)
	serializedData, err := serializer.Serialize("shoes", shoe)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize shoe: %w", err)
	}

	log.Println("Schema registration and serialization completed successfully")
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
	brokers := []string{cfg.Kafka.BootstrapServers}
	producer, err := sarama.NewAsyncProducer(brokers, saramaConfig)
	if err != nil {
		return fmt.Errorf("failed to create async producer: %w", err)
	}
	defer func() {
		if err := producer.Close(); err != nil {
			log.Printf("Error closing producer: %v", err)
		}
	}()

	// Create the message
	message := &sarama.ProducerMessage{
		Topic: cfg.Kafka.Topic,
		Value: sarama.ByteEncoder(data),
		Headers: []sarama.RecordHeader{
			{
				Key:   []byte("content-type"),
				Value: []byte("application/x-protobuf"),
			},
			{
				Key:   []byte("schema-type"),
				Value: []byte("protobuf-shoe"),
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
	case err := <-producer.Errors():
		return fmt.Errorf("❌ failed to produce message: %w", err.Err)
	case <-time.After(30 * time.Second):
		return fmt.Errorf("⏰ timeout waiting for message production")
	}
}
