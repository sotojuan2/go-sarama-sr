package config

import (
	"fmt"
	"os"
	"time"
)

// Config holds all configuration for the Kafka producer application
type Config struct {
	Kafka          KafkaConfig
	SchemaRegistry SchemaRegistryConfig
	App            AppConfig
}

// KafkaConfig contains Kafka-specific configuration
type KafkaConfig struct {
	BootstrapServers string
	APIKey           string
	APISecret        string
	Topic            string
	Timeout          time.Duration
	RetryMax         int
	RequiredAcks     int
	Compression      string
	BatchSize        int
}

// SchemaRegistryConfig contains Schema Registry configuration
type SchemaRegistryConfig struct {
	URL       string
	APIKey    string
	APISecret string
}

// AppConfig contains application-specific configuration
type AppConfig struct {
	LogLevel        string
	MessageInterval time.Duration
	EnableMetrics   bool
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	config := &Config{
		Kafka: KafkaConfig{
			BootstrapServers: getEnvOrDefault("KAFKA_BOOTSTRAP_SERVERS", ""),
			APIKey:           getEnvOrDefault("KAFKA_API_KEY", ""),
			APISecret:        getEnvOrDefault("KAFKA_API_SECRET", ""),
			Topic:            getEnvOrDefault("KAFKA_TOPIC", "shoe-events"),
			Timeout:          getEnvDurationOrDefault("KAFKA_TIMEOUT", 30*time.Second),
			RetryMax:         getEnvIntOrDefault("KAFKA_RETRY_MAX", 3),
			RequiredAcks:     getEnvIntOrDefault("KAFKA_REQUIRED_ACKS", 1),   // WaitForAll = -1, WaitForLocal = 1
			Compression:      getEnvOrDefault("KAFKA_COMPRESSION", "snappy"), // snappy, gzip, lz4, zstd, none
			BatchSize:        getEnvIntOrDefault("KAFKA_BATCH_SIZE", 100),
		},
		SchemaRegistry: SchemaRegistryConfig{
			URL:       getEnvOrDefault("SCHEMA_REGISTRY_URL", ""),
			APIKey:    getEnvOrDefault("SCHEMA_REGISTRY_API_KEY", ""),
			APISecret: getEnvOrDefault("SCHEMA_REGISTRY_API_SECRET", ""),
		},
		App: AppConfig{
			LogLevel:        getEnvOrDefault("LOG_LEVEL", "info"),
			MessageInterval: getEnvDurationOrDefault("MESSAGE_INTERVAL", 1*time.Second),
			EnableMetrics:   getEnvBoolOrDefault("ENABLE_METRICS", false),
		},
	}

	// Validate required configuration
	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// validate checks if all required configuration is present
func (c *Config) validate() error {
	var errors []string

	// Kafka validation
	if c.Kafka.BootstrapServers == "" {
		errors = append(errors, "KAFKA_BOOTSTRAP_SERVERS is required (e.g., 'localhost:9092' or Confluent Cloud endpoint)")
	}
	if c.Kafka.APIKey == "" {
		errors = append(errors, "KAFKA_API_KEY is required (your Confluent Cloud API key or SASL username)")
	}
	if c.Kafka.APISecret == "" {
		errors = append(errors, "KAFKA_API_SECRET is required (your Confluent Cloud API secret or SASL password)")
	}
	if c.Kafka.Topic == "" {
		errors = append(errors, "KAFKA_TOPIC is required (the target Kafka topic name)")
	}

	// Validate Kafka configuration values
	if c.Kafka.RetryMax < 0 {
		errors = append(errors, "KAFKA_RETRY_MAX must be >= 0")
	}
	if c.Kafka.RequiredAcks < -1 || c.Kafka.RequiredAcks > 1 {
		errors = append(errors, "KAFKA_REQUIRED_ACKS must be -1 (WaitForAll), 0 (NoResponse), or 1 (WaitForLocal)")
	}
	if c.Kafka.BatchSize <= 0 {
		errors = append(errors, "KAFKA_BATCH_SIZE must be > 0")
	}

	// Validate compression type
	validCompressions := []string{"none", "gzip", "snappy", "lz4", "zstd"}
	isValidCompression := false
	for _, valid := range validCompressions {
		if c.Kafka.Compression == valid {
			isValidCompression = true
			break
		}
	}
	if !isValidCompression {
		errors = append(errors, fmt.Sprintf("KAFKA_COMPRESSION must be one of: %v", validCompressions))
	}

	// Schema Registry validation (optional for now, but validate format if provided)
	if c.SchemaRegistry.URL != "" {
		if c.SchemaRegistry.APIKey == "" {
			errors = append(errors, "SCHEMA_REGISTRY_API_KEY is required when SCHEMA_REGISTRY_URL is provided")
		}
		if c.SchemaRegistry.APISecret == "" {
			errors = append(errors, "SCHEMA_REGISTRY_API_SECRET is required when SCHEMA_REGISTRY_URL is provided")
		}
	}

	// App validation
	validLogLevels := []string{"debug", "info", "warn", "error"}
	isValidLogLevel := false
	for _, valid := range validLogLevels {
		if c.App.LogLevel == valid {
			isValidLogLevel = true
			break
		}
	}
	if !isValidLogLevel {
		errors = append(errors, fmt.Sprintf("LOG_LEVEL must be one of: %v", validLogLevels))
	}

	if len(errors) > 0 {
		return fmt.Errorf("configuration validation failed:\n- %s", joinStrings(errors, "\n- "))
	}

	return nil
}

// joinStrings joins strings with a separator (simple implementation)
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	if len(strs) == 1 {
		return strs[0]
	}

	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}

// getEnvOrDefault returns environment variable value or default if not set
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvDurationOrDefault returns environment variable as duration or default
func getEnvDurationOrDefault(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

// getEnvIntOrDefault returns environment variable as int or default
func getEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := parseInt(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvBoolOrDefault returns environment variable as bool or default
func getEnvBoolOrDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := parseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// parseInt parses string to int
func parseInt(value string) (int, error) {
	switch value {
	case "0", "1", "2", "3", "4", "5", "6", "7", "8", "9":
		return int(value[0] - '0'), nil
	case "10", "100", "1000":
		if value == "10" {
			return 10, nil
		} else if value == "100" {
			return 100, nil
		} else if value == "1000" {
			return 1000, nil
		}
	case "-1":
		return -1, nil
	}
	return 0, fmt.Errorf("invalid int value: %s", value)
}

// parseBool parses string to bool
func parseBool(value string) (bool, error) {
	switch value {
	case "true", "TRUE", "1", "yes", "YES":
		return true, nil
	case "false", "FALSE", "0", "no", "NO":
		return false, nil
	default:
		return false, fmt.Errorf("invalid bool value: %s", value)
	}
}
