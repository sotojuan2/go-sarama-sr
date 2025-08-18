package test

import (
	"os"
	"testing"

	"github.com/go-sarama-sr/producer/internal/config"
	"github.com/go-sarama-sr/producer/pkg/schemaregistry"
)

func TestNewClient(t *testing.T) {
	// Set test environment variables for Schema Registry
	os.Setenv("SCHEMA_REGISTRY_URL", "https://test-sr.confluent.cloud")
	os.Setenv("SCHEMA_REGISTRY_API_KEY", "test-sr-key")
	os.Setenv("SCHEMA_REGISTRY_API_SECRET", "test-sr-secret")
	
	// Also set required Kafka variables to satisfy config validation
	os.Setenv("KAFKA_BOOTSTRAP_SERVERS", "test-cluster.confluent.cloud:9092")
	os.Setenv("KAFKA_API_KEY", "test-key")
	os.Setenv("KAFKA_API_SECRET", "test-secret")
	os.Setenv("KAFKA_TOPIC", "test-topic")
	
	defer func() {
		// Clean up
		envVars := []string{
			"SCHEMA_REGISTRY_URL", "SCHEMA_REGISTRY_API_KEY", "SCHEMA_REGISTRY_API_SECRET",
			"KAFKA_BOOTSTRAP_SERVERS", "KAFKA_API_KEY", "KAFKA_API_SECRET", "KAFKA_TOPIC",
		}
		for _, env := range envVars {
			os.Unsetenv(env)
		}
	}()

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test Schema Registry client creation
	client, err := schemaregistry.NewClient(cfg)
	if err != nil {
		t.Fatalf("Failed to create Schema Registry client: %v", err)
	}

	// Verify client is not nil
	if client == nil {
		t.Fatal("Expected client to be non-nil")
	}

	// Verify we can get the underlying client
	underlyingClient := client.GetClient()
	if underlyingClient == nil {
		t.Fatal("Expected underlying client to be non-nil")
	}

	// Verify we can get the config
	clientConfig := client.GetConfig()
	if clientConfig == nil {
		t.Fatal("Expected client config to be non-nil")
	}

	// Verify configuration values
	if clientConfig.SchemaRegistry.URL != "https://test-sr.confluent.cloud" {
		t.Errorf("Expected URL to be 'https://test-sr.confluent.cloud', got %s", clientConfig.SchemaRegistry.URL)
	}

	// Test Close method
	err = client.Close()
	if err != nil {
		t.Errorf("Failed to close client: %v", err)
	}

	t.Logf("✅ Schema Registry client creation test passed")
}

func TestNewClientValidation(t *testing.T) {
	testCases := []struct {
		name            string
		url             string
		apiKey          string
		apiSecret       string
		expectedError   string
	}{
		{
			name:          "missing URL",
			url:           "",
			apiKey:        "test-key-valid",
			apiSecret:     "test-secret-valid",
			expectedError: "SCHEMA_REGISTRY_URL is required",
		},
		{
			name:          "missing API key",
			url:           "https://test-sr.confluent.cloud",
			apiKey:        "",
			apiSecret:     "test-secret-valid",
			expectedError: "SCHEMA_REGISTRY_API_KEY is required",
		},
		{
			name:          "missing API secret",
			url:           "https://test-sr.confluent.cloud",
			apiKey:        "test-key-valid",
			apiSecret:     "",
			expectedError: "SCHEMA_REGISTRY_API_SECRET is required",
		},
		{
			name:          "invalid URL format - no protocol",
			url:           "test-sr.confluent.cloud",
			apiKey:        "test-key-valid",
			apiSecret:     "test-secret-valid",
			expectedError: "URL must start with http:// or https://",
		},
		{
			name:          "invalid URL format - spaces",
			url:           "https://test sr.confluent.cloud",
			apiKey:        "test-key-valid",
			apiSecret:     "test-secret-valid",
			expectedError: "URL cannot contain spaces",
		},
		{
			name:          "API key too short",
			url:           "https://test-sr.confluent.cloud",
			apiKey:        "short",
			apiSecret:     "test-secret-valid",
			expectedError: "API key must be at least",
		},
		{
			name:          "API secret too short",
			url:           "https://test-sr.confluent.cloud",
			apiKey:        "test-key-valid",
			apiSecret:     "short",
			expectedError: "API secret must be at least",
		},
		{
			name:          "API key with whitespace",
			url:           "https://test-sr.confluent.cloud",
			apiKey:        " test-key-valid ",
			apiSecret:     "test-secret-valid",
			expectedError: "API key cannot contain leading or trailing whitespace",
		},
		{
			name:          "API secret with whitespace",
			url:           "https://test-sr.confluent.cloud",
			apiKey:        "test-key-valid",
			apiSecret:     " test-secret-valid ",
			expectedError: "API secret cannot contain leading or trailing whitespace",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set up minimal config
			cfg := &config.Config{
				SchemaRegistry: config.SchemaRegistryConfig{
					URL:       tc.url,
					APIKey:    tc.apiKey,
					APISecret: tc.apiSecret,
				},
			}

			_, err := schemaregistry.NewClient(cfg)
			if err == nil {
				t.Errorf("Expected error for test case '%s', but got none", tc.name)
			}

			if err != nil && !containsString(err.Error(), tc.expectedError) {
				t.Errorf("Expected error to contain '%s', but got: %v", tc.expectedError, err)
			}
		})
	}
}

// TestNewClientWithValidCredentials tests creation with properly formatted credentials
func TestNewClientWithValidCredentials(t *testing.T) {
	testCases := []struct {
		name      string
		url       string
		apiKey    string
		apiSecret string
	}{
		{
			name:      "https URL with valid credentials",
			url:       "https://test-sr.confluent.cloud",
			apiKey:    "ABCD1234EFGH5678",  // Confluent Cloud format
			apiSecret: "abcdefghij1234567890abcdefghij1234567890abcdefghij1234567890",
		},
		{
			name:      "http URL (for local testing)",
			url:       "http://localhost:8081",
			apiKey:    "local-test-key-123",
			apiSecret: "local-test-secret-456789",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &config.Config{
				SchemaRegistry: config.SchemaRegistryConfig{
					URL:       tc.url,
					APIKey:    tc.apiKey,
					APISecret: tc.apiSecret,
				},
			}

			// This should pass validation but will fail on actual connection
			// since we're using test credentials
			_, err := schemaregistry.NewClient(cfg)
			
			// We expect this to fail at authentication step, not validation
			if err != nil {
				// Check that it's not a validation error
				if containsString(err.Error(), "configuration validation failed") {
					t.Errorf("Validation failed unexpectedly: %v", err)
				}
				// Other errors are expected (network, auth, etc.)
				t.Logf("Expected error during connection (not validation): %v", err)
			}
		})
	}
}
