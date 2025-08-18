package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/IBM/sarama"
)

// KafkaProducer wraps the Sarama async producer
type KafkaProducer struct {
	producer sarama.AsyncProducer
	topic    string
}

// NewKafkaProducer creates a new Kafka producer with Confluent Cloud configuration
func NewKafkaProducer(brokers []string, username, password, topic string) (*KafkaProducer, error) {
	config := sarama.NewConfig()

	// Configure SASL_SSL for Confluent Cloud
	config.Net.SASL.Enable = true
	config.Net.SASL.User = username
	config.Net.SASL.Password = password
	config.Net.SASL.Mechanism = sarama.SASLTypePlaintext
	config.Net.TLS.Enable = true

	// Configure producer settings
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	config.Producer.RequiredAcks = sarama.WaitForAll // Wait for all replicas
	config.Producer.Retry.Max = 5
	config.Producer.Compression = sarama.CompressionSnappy
	config.Version = sarama.V3_5_0_0 // Use latest supported version

	// Create async producer
	producer, err := sarama.NewAsyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}

	kp := &KafkaProducer{
		producer: producer,
		topic:    topic,
	}

	// Start goroutines to handle successes and errors
	go kp.handleSuccesses()
	go kp.handleErrors()

	return kp, nil
}

// handleSuccesses processes successful message deliveries
func (kp *KafkaProducer) handleSuccesses() {
	for msg := range kp.producer.Successes() {
		log.Printf("✅ Message sent successfully to topic %s partition %d offset %d",
			msg.Topic, msg.Partition, msg.Offset)
	}
}

// handleErrors processes producer errors
func (kp *KafkaProducer) handleErrors() {
	for err := range kp.producer.Errors() {
		log.Printf("❌ Failed to send message to topic %s: %v", err.Msg.Topic, err.Err)
	}
}

// SendMessage sends a message to the configured Kafka topic
func (kp *KafkaProducer) SendMessage(key, value string) {
	msg := &sarama.ProducerMessage{
		Topic: kp.topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.StringEncoder(value),
	}

	kp.producer.Input() <- msg
	log.Printf("📤 Sent message with key '%s' to topic '%s'", key, kp.topic)
}

// Close gracefully shuts down the producer
func (kp *KafkaProducer) Close() error {
	log.Println("🔄 Closing Kafka producer...")
	return kp.producer.Close()
}

func main() {
	log.Println("🚀 Starting Kafka Producer for Task 2 - Basic Connectivity Test")

	// For now, use placeholder values - will be replaced with env vars in Task 3
	// These are typical Confluent Cloud values format
	brokers := []string{"pkc-example.us-west-2.aws.confluent.cloud:9092"}
	username := "placeholder-api-key"
	password := "placeholder-api-secret"
	topic := "shoe-events"

	log.Printf("📋 Configuration:")
	log.Printf("   Brokers: %v", brokers)
	log.Printf("   Topic: %s", topic)
	log.Printf("   Username: %s", username)
	log.Printf("   (This is a connectivity test with placeholder credentials)")

	// Create producer
	producer, err := NewKafkaProducer(brokers, username, password, topic)
	if err != nil {
		log.Fatalf("❌ Failed to create Kafka producer: %v", err)
	}
	defer producer.Close()

	// Set up graceful shutdown
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)

	// Send a test message
	log.Println("📨 Sending test message...")
	producer.SendMessage("test-key-1", "Hello Confluent Cloud from Sarama! This is a connectivity test.")

	// Wait a bit for async delivery
	log.Println("⏱️  Waiting for message delivery confirmation...")
	time.Sleep(5 * time.Second)

	// Wait for shutdown signal
	log.Println("✨ Test message sent. Press Ctrl+C to shutdown...")
	<-sigterm

	log.Println("✅ Task 2 completed - Basic Sarama connectivity implementation finished")
}
