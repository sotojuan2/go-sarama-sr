package registry

import (
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry"
	"github.com/go-sarama-sr/producer/internal/config"
)

// Client wraps the Schema Registry client
type Client struct {
	client schemaregistry.Client
	config *config.Config
}

// NewClient creates a new Schema Registry client
func NewClient(cfg *config.Config) (*Client, error) {
	// Create config with authentication
	config := schemaregistry.NewConfig(cfg.SchemaRegistry.URL)
	config.BasicAuthUserInfo = cfg.SchemaRegistry.APIKey + ":" + cfg.SchemaRegistry.APISecret

	client, err := schemaregistry.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create schema registry client: %w", err)
	}

	return &Client{
		client: client,
		config: cfg,
	}, nil
}

// GetClient returns the underlying Schema Registry client
func (c *Client) GetClient() schemaregistry.Client {
	return c.client
}

// HealthCheck verifies connectivity to Schema Registry
func (c *Client) HealthCheck() error {
	// Try to list subjects (this will fail if we can't connect or authenticate)
	_, err := c.client.GetAllSubjects()
	if err != nil {
		return fmt.Errorf("Schema Registry health check failed: %w", err)
	}
	return nil
}
