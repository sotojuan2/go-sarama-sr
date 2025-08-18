package main

import (
	"log"

	"github.com/go-sarama-sr/producer/internal/config"
	"github.com/go-sarama-sr/producer/pkg/kafka"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Could not load .env file: %v", err)
		log.Println("Continuing with system environment variables...")
	}

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("❌ Failed to load configuration: %v", err)
	}

	log.Printf("🚀 Starting Kafka producer for topic: %s", cfg.Kafka.Topic)
	log.Printf("📡 Connecting to: %s", cfg.Kafka.BootstrapServers)

	// Create Kafka producer
	producer, err := kafka.NewProducer(cfg)
	if err != nil {
		log.Fatalf("❌ Failed to create Kafka producer: %v", err)
	}
	defer func() {
		if err := producer.Close(); err != nil {
			log.Printf("❌ Error closing producer: %v", err)
		}
	}()

	// Send a test message
	testMessage := "Hello from go-sarama-sr producer! This is a test message to validate connectivity."

	log.Printf("📤 Sending test message...")
	if err := producer.SendMessage(testMessage); err != nil {
		log.Fatalf("❌ Failed to send message: %v", err)
	}

	log.Printf("✅ Task 2 completed successfully! Basic Sarama producer connectivity established.")
	log.Printf("🎉 Producer can successfully connect to Confluent Cloud and send messages.")
}
