package test

import (
	"os"
	"testing"
	"time"

	"github.com/go-sarama-sr/producer/internal/config"
)

func TestConfigLoading(t *testing.T) {
	// Set test environment variables
	os.Setenv("KAFKA_BOOTSTRAP_SERVERS", "test-cluster.confluent.cloud:9092")
	os.Setenv("KAFKA_API_KEY", "test-api-key")
	os.Setenv("KAFKA_API_SECRET", "test-api-secret")
	os.Setenv("KAFKA_TOPIC", "test-topic")
	os.Setenv("KAFKA_RETRY_MAX", "5")
	os.Setenv("KAFKA_REQUIRED_ACKS", "-1")
	os.Setenv("KAFKA_COMPRESSION", "gzip")
	os.Setenv("KAFKA_BATCH_SIZE", "200")
	os.Setenv("SCHEMA_REGISTRY_URL", "https://test-sr.confluent.cloud")
	os.Setenv("SCHEMA_REGISTRY_API_KEY", "sr-test-key")
	os.Setenv("SCHEMA_REGISTRY_API_SECRET", "sr-test-secret")
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("MESSAGE_INTERVAL", "2s")
	os.Setenv("ENABLE_METRICS", "true")
	
	// Clean up after test
	defer func() {
		envVars := []string{
			"KAFKA_BOOTSTRAP_SERVERS", "KAFKA_API_KEY", "KAFKA_API_SECRET", "KAFKA_TOPIC",
			"KAFKA_RETRY_MAX", "KAFKA_REQUIRED_ACKS", "KAFKA_COMPRESSION", "KAFKA_BATCH_SIZE",
			"SCHEMA_REGISTRY_URL", "SCHEMA_REGISTRY_API_KEY", "SCHEMA_REGISTRY_API_SECRET",
			"LOG_LEVEL", "MESSAGE_INTERVAL", "ENABLE_METRICS",
		}
		for _, env := range envVars {
			os.Unsetenv(env)
		}
	}()

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Test Kafka configuration
	if cfg.Kafka.BootstrapServers != "test-cluster.confluent.cloud:9092" {
		t.Errorf("Expected bootstrap servers to be 'test-cluster.confluent.cloud:9092', got %s", cfg.Kafka.BootstrapServers)
	}
	if cfg.Kafka.APIKey != "test-api-key" {
		t.Errorf("Expected API key to be 'test-api-key', got %s", cfg.Kafka.APIKey)
	}
	if cfg.Kafka.APISecret != "test-api-secret" {
		t.Errorf("Expected API secret to be 'test-api-secret', got %s", cfg.Kafka.APISecret)
	}
	if cfg.Kafka.Topic != "test-topic" {
		t.Errorf("Expected topic to be 'test-topic', got %s", cfg.Kafka.Topic)
	}
	if cfg.Kafka.RetryMax != 5 {
		t.Errorf("Expected retry max to be 5, got %d", cfg.Kafka.RetryMax)
	}
	if cfg.Kafka.RequiredAcks != -1 {
		t.Errorf("Expected required acks to be -1, got %d", cfg.Kafka.RequiredAcks)
	}
	if cfg.Kafka.Compression != "gzip" {
		t.Errorf("Expected compression to be 'gzip', got %s", cfg.Kafka.Compression)
	}
	if cfg.Kafka.BatchSize != 200 {
		t.Errorf("Expected batch size to be 200, got %d", cfg.Kafka.BatchSize)
	}

	// Test Schema Registry configuration
	if cfg.SchemaRegistry.URL != "https://test-sr.confluent.cloud" {
		t.Errorf("Expected Schema Registry URL to be 'https://test-sr.confluent.cloud', got %s", cfg.SchemaRegistry.URL)
	}
	if cfg.SchemaRegistry.APIKey != "sr-test-key" {
		t.Errorf("Expected Schema Registry API key to be 'sr-test-key', got %s", cfg.SchemaRegistry.APIKey)
	}
	if cfg.SchemaRegistry.APISecret != "sr-test-secret" {
		t.Errorf("Expected Schema Registry API secret to be 'sr-test-secret', got %s", cfg.SchemaRegistry.APISecret)
	}

	// Test App configuration
	if cfg.App.LogLevel != "debug" {
		t.Errorf("Expected log level to be 'debug', got %s", cfg.App.LogLevel)
	}
	if cfg.App.MessageInterval != 2*time.Second {
		t.Errorf("Expected message interval to be 2s, got %v", cfg.App.MessageInterval)
	}
	if !cfg.App.EnableMetrics {
		t.Errorf("Expected enable metrics to be true, got %v", cfg.App.EnableMetrics)
	}

	t.Logf("✅ Enhanced configuration test passed - all values loaded correctly")
}

func TestConfigValidation(t *testing.T) {
	// Test missing required fields and invalid values
	testCases := []struct {
		name        string
		setupEnv    func()
		expectError bool
		errorContains string
	}{
		{
			name: "missing bootstrap servers",
			setupEnv: func() {
				os.Unsetenv("KAFKA_BOOTSTRAP_SERVERS")
				os.Setenv("KAFKA_API_KEY", "test-key")
				os.Setenv("KAFKA_API_SECRET", "test-secret")
				os.Setenv("KAFKA_TOPIC", "test-topic")
			},
			expectError: true,
			errorContains: "KAFKA_BOOTSTRAP_SERVERS is required",
		},
		{
			name: "missing API key",
			setupEnv: func() {
				os.Setenv("KAFKA_BOOTSTRAP_SERVERS", "test-cluster.confluent.cloud:9092")
				os.Unsetenv("KAFKA_API_KEY")
				os.Setenv("KAFKA_API_SECRET", "test-secret")
				os.Setenv("KAFKA_TOPIC", "test-topic")
			},
			expectError: true,
			errorContains: "KAFKA_API_KEY is required",
		},
		{
			name: "invalid compression",
			setupEnv: func() {
				os.Setenv("KAFKA_BOOTSTRAP_SERVERS", "test-cluster.confluent.cloud:9092")
				os.Setenv("KAFKA_API_KEY", "test-key")
				os.Setenv("KAFKA_API_SECRET", "test-secret")
				os.Setenv("KAFKA_TOPIC", "test-topic")
				os.Setenv("KAFKA_COMPRESSION", "invalid")
			},
			expectError: true,
			errorContains: "KAFKA_COMPRESSION must be one of",
		},
		{
			name: "invalid log level",
			setupEnv: func() {
				os.Setenv("KAFKA_BOOTSTRAP_SERVERS", "test-cluster.confluent.cloud:9092")
				os.Setenv("KAFKA_API_KEY", "test-key")
				os.Setenv("KAFKA_API_SECRET", "test-secret")
				os.Setenv("KAFKA_TOPIC", "test-topic")
				os.Setenv("LOG_LEVEL", "invalid")
			},
			expectError: true,
			errorContains: "LOG_LEVEL must be one of",
		},
		{
			name: "schema registry URL without credentials",
			setupEnv: func() {
				os.Setenv("KAFKA_BOOTSTRAP_SERVERS", "test-cluster.confluent.cloud:9092")
				os.Setenv("KAFKA_API_KEY", "test-key")
				os.Setenv("KAFKA_API_SECRET", "test-secret")
				os.Setenv("KAFKA_TOPIC", "test-topic")
				os.Setenv("SCHEMA_REGISTRY_URL", "https://test-sr.confluent.cloud")
				// Missing API key and secret
			},
			expectError: true,
			errorContains: "SCHEMA_REGISTRY_API_KEY is required",
		},
		{
			name: "valid configuration with all optional fields",
			setupEnv: func() {
				os.Setenv("KAFKA_BOOTSTRAP_SERVERS", "test-cluster.confluent.cloud:9092")
				os.Setenv("KAFKA_API_KEY", "test-key")
				os.Setenv("KAFKA_API_SECRET", "test-secret")
				os.Setenv("KAFKA_TOPIC", "test-topic")
				os.Setenv("KAFKA_COMPRESSION", "snappy")
				os.Setenv("LOG_LEVEL", "info")
				os.Setenv("SCHEMA_REGISTRY_URL", "https://test-sr.confluent.cloud")
				os.Setenv("SCHEMA_REGISTRY_API_KEY", "sr-key")
				os.Setenv("SCHEMA_REGISTRY_API_SECRET", "sr-secret")
			},
			expectError: false,
		},
		{
			name: "valid configuration minimal",
			setupEnv: func() {
				// Clear all non-required env vars
				envVars := []string{
					"KAFKA_RETRY_MAX", "KAFKA_REQUIRED_ACKS", "KAFKA_COMPRESSION", "KAFKA_BATCH_SIZE",
					"SCHEMA_REGISTRY_URL", "SCHEMA_REGISTRY_API_KEY", "SCHEMA_REGISTRY_API_SECRET",
					"LOG_LEVEL", "MESSAGE_INTERVAL", "ENABLE_METRICS",
				}
				for _, env := range envVars {
					os.Unsetenv(env)
				}
				// Set only required vars
				os.Setenv("KAFKA_BOOTSTRAP_SERVERS", "test-cluster.confluent.cloud:9092")
				os.Setenv("KAFKA_API_KEY", "test-key")
				os.Setenv("KAFKA_API_SECRET", "test-secret")
				os.Setenv("KAFKA_TOPIC", "test-topic")
			},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupEnv()
			
			_, err := config.LoadConfig()
			if tc.expectError && err == nil {
				t.Errorf("Expected error for test case '%s', but got none", tc.name)
			}
			if !tc.expectError && err != nil {
				t.Errorf("Expected no error for test case '%s', but got: %v", tc.name, err)
			}
			if tc.expectError && err != nil && tc.errorContains != "" {
				if !containsString(err.Error(), tc.errorContains) {
					t.Errorf("Expected error to contain '%s', but got: %v", tc.errorContains, err)
				}
			}
		})
	}
}

// containsString checks if a string contains a substring (simple implementation)
func containsString(str, substr string) bool {
	return len(str) >= len(substr) && findSubstring(str, substr) >= 0
}

// findSubstring finds the index of a substring in a string
func findSubstring(str, substr string) int {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
