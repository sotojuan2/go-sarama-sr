package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/IBM/sarama"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde/protobuf"
	pb "github.com/go-sarama-sr/producer/pb"
	"github.com/go-sarama-sr/producer/pkg/config"
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
	srClientWrapper, err := srClient.NewClient(&cfg.SchemaRegistry)
	if err != nil {
		log.Fatalf("Error creating Schema Registry client: %v", err)
	}

	// Subtask 5.1: Create hardcoded Shoe object
	shoe := createHardcodedShoe()
	log.Printf("Successfully created shoe object: %+v", shoe)

	// Subtask 5.2 & 5.3: Serialize with Schema Registry (schema auto-registration)
	serializedData, err := serializeShoeWithSchemaRegistry(srClientWrapper, shoe)
	if err != nil {
		log.Fatalf("Error serializing shoe: %v", err)
	}
	log.Printf("Successfully serialized shoe data: %d bytes", len(serializedData))

	// Subtask 5.4: Produce message with Sarama
	err = produceMessageToKafka(cfg, serializedData)
	if err != nil {
		log.Fatalf("Error producing message to Kafka: %v", err)
	}

	log.Println("Producer completed successfully - Message sent to Kafka!")
}

// serializeShoeWithSchemaRegistry serializa el objeto Shoe usando Schema Registry
func serializeShoeWithSchemaRegistry(client *srClient.Client, shoe *pb.Shoe) ([]byte, error) {
	log.Println("Serializing shoe with Schema Registry...")

	// Configurar serializer de Protobuf
	serializer, err := protobuf.NewSerializer(client.GetClient(), serde.ValueSerde, protobuf.NewSerializerConfig())
	if err != nil {
		return nil, fmt.Errorf("failed to create protobuf serializer: %w", err)
	}

	// Serializar el objeto (el schema se registra automáticamente si no existe)
	ctx := context.Background()
	serializedData, err := serializer.Serialize(ctx, "shoes", shoe)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize shoe: %w", err)
	}

	log.Println("Schema registration and serialization completed successfully")
	return serializedData, nil
}

// produceMessageToKafka produce el mensaje serializado a Kafka usando Sarama
func produceMessageToKafka(cfg *config.Config, data []byte) error {
	log.Println("Setting up Kafka producer...")

	// Configurar Sarama
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true

	// Configuración SASL para Confluent Cloud
	config.Net.SASL.Enable = true
	config.Net.SASL.Mechanism = sarama.SASLTypePlaintext
	config.Net.SASL.User = cfg.Kafka.Username
	config.Net.SASL.Password = cfg.Kafka.Password
	config.Net.TLS.Enable = true

	// Crear producer asíncrono
	producer, err := sarama.NewAsyncProducer(cfg.Kafka.Brokers, config)
	if err != nil {
		return fmt.Errorf("failed to create async producer: %w", err)
	}
	defer producer.Close()

	// Crear el mensaje
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

	// Enviar mensaje
	producer.Input() <- message

	// Manejar respuesta
	select {
	case success := <-producer.Successes():
		log.Printf("Message produced successfully to partition %d at offset %d",
			success.Partition, success.Offset)
		return nil
	case err := <-producer.Errors():
		return fmt.Errorf("failed to produce message: %w", err.Err)
	case <-time.After(30 * time.Second):
		return fmt.Errorf("timeout waiting for message production")
	}
}
