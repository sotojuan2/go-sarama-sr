package main

import (
	"log"

	"github.com/go-sarama-sr/producer/internal/config"
	"github.com/go-sarama-sr/producer/pkg/schemaregistry"
	"github.com/joho/godotenv"
)

func main() {
	log.Println("🚀 Testing Schema Registry connectivity - Subtask 4.2")

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("⚠️  Warning: .env file not found: %v", err)
	}

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("❌ Failed to load configuration: %v", err)
	}

	log.Printf("📋 Schema Registry Configuration:")
	log.Printf("   URL: %s", cfg.SchemaRegistry.URL)
	log.Printf("   API Key: %s", cfg.SchemaRegistry.APIKey)
	log.Printf("   (API Secret configured: %t)", cfg.SchemaRegistry.APISecret != "")

	// Create Schema Registry client
	client, err := schemaregistry.NewClient(cfg)
	if err != nil {
		log.Fatalf("❌ Failed to create Schema Registry client: %v", err)
	}
	defer client.Close()

	// Test connection
	log.Println("🔗 Testing Schema Registry connectivity...")
	if err := client.TestConnection(); err != nil {
		log.Fatalf("❌ Schema Registry connection test failed: %v", err)
	}

	log.Println("✅ Subtask 4.2 completed successfully!")
	log.Println("🎉 Schema Registry client can connect and authenticate with Confluent Cloud")
}
