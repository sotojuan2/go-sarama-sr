package main

import (
	"log"
	"os"

	"github.com/go-sarama-sr/producer/internal/config"
	"github.com/go-sarama-sr/producer/pkg/schemaregistry"
)

func main() {
	log.Println("🚀 Testing Enhanced Schema Registry Authentication...")

	// Load configuration directly from environment (for testing)
	cfg := &config.Config{
		SchemaRegistry: config.SchemaRegistryConfig{
			URL:       os.Getenv("SCHEMA_REGISTRY_URL"),
			APIKey:    os.Getenv("SCHEMA_REGISTRY_API_KEY"),
			APISecret: os.Getenv("SCHEMA_REGISTRY_API_SECRET"),
		},
	}

	log.Printf("🔧 Configuration loaded:")
	log.Printf("  - Schema Registry URL: %s", cfg.SchemaRegistry.URL)
	log.Printf("  - API Key: %s", cfg.SchemaRegistry.APIKey)
	log.Printf("  - API Secret: [%d characters]", len(cfg.SchemaRegistry.APISecret))

	// Create Schema Registry client with enhanced authentication
	client, err := schemaregistry.NewClient(cfg)
	if err != nil {
		log.Fatalf("❌ Failed to create Schema Registry client: %v", err)
	}
	defer client.Close()

	log.Println("✅ Schema Registry client created successfully with enhanced authentication!")

	// Test connection
	if err := client.TestConnection(); err != nil {
		log.Fatalf("❌ Connection test failed: %v", err)
	}

	log.Println("🎉 Enhanced authentication test completed successfully!")
	log.Println("🔐 All authentication validations and connectivity tests passed!")

	// Exit cleanly
	os.Exit(0)
}
