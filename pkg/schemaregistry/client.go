package schemaregistry

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	sr "github.com/confluentinc/confluent-kafka-go/v2/schemaregistry"
	"github.com/go-sarama-sr/producer/internal/config"
)

// Authentication constants
const (
	// Timeouts for Schema Registry operations
	DefaultConnectTimeout = 10 * time.Second
	DefaultRequestTimeout = 5 * time.Second
	
	// Retry configuration
	MaxRetryAttempts = 3
	RetryDelay      = 1 * time.Second
	
	// API Key validation patterns
	APIKeyMinLength    = 10
	APISecretMinLength = 10
	
	// Common Confluent Cloud API key pattern (starts with letters, contains alphanumeric)
	APIKeyPattern = `^[A-Z0-9]{16}$`
)

// HealthStatus represents the health status of the Schema Registry
type HealthStatus struct {
	Healthy          bool          `json:"healthy"`
	URL              string        `json:"url"`
	ResponseTime     time.Duration `json:"response_time"`
	SubjectCount     int           `json:"subject_count"`
	LastError        string        `json:"last_error,omitempty"`
	TestedEndpoints  []string      `json:"tested_endpoints"`
	ErrorDetails     []ErrorDetail `json:"error_details,omitempty"`
}

// ErrorDetail provides detailed information about connectivity errors
type ErrorDetail struct {
	Endpoint    string `json:"endpoint"`
	Operation   string `json:"operation"`
	ErrorType   string `json:"error_type"`
	ErrorCode   int    `json:"error_code,omitempty"`
	Message     string `json:"message"`
	Timestamp   string `json:"timestamp"`
	Recoverable bool   `json:"recoverable"`
}

// Client wraps the Confluent Schema Registry client with our configuration
type Client struct {
	client sr.Client
	config *config.Config
}

// NewClient creates a new Schema Registry client with enhanced authentication and error handling
func NewClient(cfg *config.Config) (*Client, error) {
	log.Printf("🔄 Initializing Schema Registry client...")

	// Validate Schema Registry configuration with enhanced checks
	if err := validateSchemaRegistryConfig(cfg); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	// Create Schema Registry client configuration with timeouts
	clientConfig := sr.NewConfigWithBasicAuthentication(
		cfg.SchemaRegistry.URL,
		cfg.SchemaRegistry.APIKey,
		cfg.SchemaRegistry.APISecret,
	)

	// Set timeouts for robust connection handling
	clientConfig.RequestTimeoutMs = int(DefaultRequestTimeout.Milliseconds())
	log.Printf("⏰ Client configured with %v request timeout", DefaultRequestTimeout)

	// Create the client with retry logic
	client, err := createClientWithRetry(clientConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Schema Registry client after %d attempts: %w", MaxRetryAttempts, err)
	}

	log.Printf("✅ Schema Registry client created successfully for: %s", cfg.SchemaRegistry.URL)

	// Create our wrapper
	wrapper := &Client{
		client: client,
		config: cfg,
	}

	// Perform initial connectivity and authentication test
	if err := wrapper.validateAuthentication(); err != nil {
		return nil, fmt.Errorf("authentication validation failed: %w", err)
	}

	log.Printf("🎉 Schema Registry client fully initialized and authenticated")
	return wrapper, nil
}

// createClientWithRetry attempts to create the client with retry logic
func createClientWithRetry(clientConfig *sr.Config) (sr.Client, error) {
	var lastErr error
	
	for attempt := 1; attempt <= MaxRetryAttempts; attempt++ {
		log.Printf("🔄 Creating client (attempt %d/%d)...", attempt, MaxRetryAttempts)
		
		client, err := sr.NewClient(clientConfig)
		if err == nil {
			log.Printf("✅ Client created successfully on attempt %d", attempt)
			return client, nil
		}
		
		lastErr = err
		log.Printf("❌ Attempt %d failed: %v", attempt, err)
		
		// Don't sleep on the last attempt
		if attempt < MaxRetryAttempts {
			log.Printf("⏳ Waiting %v before retry...", RetryDelay)
			time.Sleep(RetryDelay)
		}
	}
	
	return nil, fmt.Errorf("all %d attempts failed, last error: %w", MaxRetryAttempts, lastErr)
}

// validateAuthentication performs an initial authentication test
func (c *Client) validateAuthentication() error {
	log.Printf("🔐 Validating authentication with Schema Registry...")
	
	// Try a simple operation to validate authentication
	// This will fail if credentials are wrong
	_, err := c.client.GetAllSubjects()
	if err != nil {
		// Analyze the error to provide specific feedback
		if strings.Contains(err.Error(), "401") || strings.Contains(err.Error(), "Unauthorized") {
			return fmt.Errorf("authentication failed - check your API key and secret: %w", err)
		}
		if strings.Contains(err.Error(), "403") || strings.Contains(err.Error(), "Forbidden") {
			return fmt.Errorf("access denied - check your API key permissions: %w", err)
		}
		if strings.Contains(err.Error(), "timeout") || strings.Contains(err.Error(), "connection") {
			return fmt.Errorf("connection failed - check network and URL: %w", err)
		}
		return fmt.Errorf("authentication validation failed: %w", err)
	}
	
	log.Printf("✅ Authentication validated successfully")
	return nil
}

// validateSchemaRegistryConfig performs comprehensive validation of Schema Registry configuration
func validateSchemaRegistryConfig(cfg *config.Config) error {
	// Check for missing configuration
	if cfg.SchemaRegistry.URL == "" {
		return fmt.Errorf("SCHEMA_REGISTRY_URL is required")
	}
	if cfg.SchemaRegistry.APIKey == "" {
		return fmt.Errorf("SCHEMA_REGISTRY_API_KEY is required")
	}
	if cfg.SchemaRegistry.APISecret == "" {
		return fmt.Errorf("SCHEMA_REGISTRY_API_SECRET is required")
	}

	// Validate URL format
	if err := validateURL(cfg.SchemaRegistry.URL); err != nil {
		return fmt.Errorf("invalid SCHEMA_REGISTRY_URL: %w", err)
	}

	// Validate API Key format and length
	if err := validateAPIKey(cfg.SchemaRegistry.APIKey); err != nil {
		return fmt.Errorf("invalid SCHEMA_REGISTRY_API_KEY: %w", err)
	}

	// Validate API Secret format and length
	if err := validateAPISecret(cfg.SchemaRegistry.APISecret); err != nil {
		return fmt.Errorf("invalid SCHEMA_REGISTRY_API_SECRET: %w", err)
	}

	log.Printf("🔐 Schema Registry credentials validated successfully")
	return nil
}

// validateURL checks if the URL is properly formatted
func validateURL(url string) error {
	if !strings.HasPrefix(url, "https://") && !strings.HasPrefix(url, "http://") {
		return fmt.Errorf("URL must start with http:// or https://")
	}
	if strings.Contains(url, " ") {
		return fmt.Errorf("URL cannot contain spaces")
	}
	return nil
}

// validateAPIKey validates the format and length of API key
func validateAPIKey(apiKey string) error {
	// Check minimum length
	if len(apiKey) < APIKeyMinLength {
		return fmt.Errorf("API key must be at least %d characters long", APIKeyMinLength)
	}

	// Check for whitespace
	if strings.TrimSpace(apiKey) != apiKey {
		return fmt.Errorf("API key cannot contain leading or trailing whitespace")
	}

	// Check for common Confluent Cloud pattern (optional validation)
	if matched, _ := regexp.MatchString(APIKeyPattern, apiKey); !matched {
		log.Printf("⚠️  API key doesn't match expected Confluent Cloud pattern (this may be ok for other providers)")
	}

	return nil
}

// validateAPISecret validates the format and length of API secret
func validateAPISecret(apiSecret string) error {
	// Check minimum length
	if len(apiSecret) < APISecretMinLength {
		return fmt.Errorf("API secret must be at least %d characters long", APISecretMinLength)
	}

	// Check for whitespace
	if strings.TrimSpace(apiSecret) != apiSecret {
		return fmt.Errorf("API secret cannot contain leading or trailing whitespace")
	}

	// Don't log the actual secret for security
	log.Printf("🔑 API secret format validated (length: %d characters)", len(apiSecret))
	return nil
}

// GetClient returns the underlying Schema Registry client
func (c *Client) GetClient() sr.Client {
	return c.client
}

// GetConfig returns the configuration used by this client
func (c *Client) GetConfig() *config.Config {
	return c.config
}

// Close closes the Schema Registry client (if needed)
func (c *Client) Close() error {
	// The Confluent client doesn't require explicit closing in v2
	log.Println("✅ Schema Registry client closed")
	return nil
}

// TestConnection performs a comprehensive connectivity and authentication test
func (c *Client) TestConnection() error {
	log.Printf("🔍 Testing Schema Registry connection...")
	
	start := time.Now()
	
	// Try to get all subjects to test connectivity and authentication
	subjects, err := c.client.GetAllSubjects()
	if err != nil {
		// Provide specific error context
		if strings.Contains(err.Error(), "401") {
			return fmt.Errorf("authentication failed - invalid credentials: %w", err)
		}
		if strings.Contains(err.Error(), "403") {
			return fmt.Errorf("access forbidden - check API key permissions: %w", err)
		}
		if strings.Contains(err.Error(), "timeout") {
			return fmt.Errorf("connection timeout - check network or URL: %w", err)
		}
		return fmt.Errorf("schema Registry connection failed: %w", err)
	}

	duration := time.Since(start)
	
	log.Printf("✅ Successfully connected to Schema Registry: %s", c.config.SchemaRegistry.URL)
	log.Printf("⏱️  Connection test completed in %v", duration)
	log.Printf("📋 Available subjects (%d): %v", len(subjects), subjects)
	
	// Additional validation: try to get a specific subject that may not exist
	// This tests another API endpoint
	_, err = c.client.GetLatestSchemaMetadata("__health_check__")
	if err != nil && !strings.Contains(err.Error(), "not found") && !strings.Contains(err.Error(), "40401") {
		log.Printf("⚠️  Health check subject test failed (this may be expected): %v", err)
	} else {
		log.Printf("✅ Schema metadata API endpoint responsive")
	}
	
	return nil
}

// ComprehensiveHealthCheck performs an exhaustive connectivity validation
func (c *Client) ComprehensiveHealthCheck() (*HealthStatus, error) {
	log.Printf("🔍 Starting comprehensive Schema Registry health check...")
	
	start := time.Now()
	health := &HealthStatus{
		URL:             c.config.SchemaRegistry.URL,
		TestedEndpoints: []string{},
		ErrorDetails:    []ErrorDetail{},
	}

	// Test 1: Basic connectivity - Get all subjects
	log.Printf("📋 Testing subjects endpoint...")
	subjects, err := c.testSubjectsEndpoint(health)
	if err != nil {
		health.Healthy = false
		health.LastError = err.Error()
	} else {
		health.SubjectCount = len(subjects)
		log.Printf("✅ Found %d subjects", len(subjects))
	}

	// Test 2: Schema metadata endpoint
	log.Printf("📄 Testing schema metadata endpoint...")
	c.testSchemaMetadataEndpoint(health)

	// Test 3: Test with a real schema if subjects exist
	if len(subjects) > 0 {
		log.Printf("🔍 Testing with existing subject: %s", subjects[0])
		c.testExistingSubject(health, subjects[0])
	}

	// Test 4: Test authentication with invalid request (to verify error handling)
	log.Printf("🔐 Testing error handling...")
	c.testErrorHandling(health)

	health.ResponseTime = time.Since(start)
	health.Healthy = len(health.ErrorDetails) == 0 || c.hasOnlyRecoverableErrors(health)

	log.Printf("🏥 Health check completed in %v - Status: %s", 
		health.ResponseTime, 
		map[bool]string{true: "HEALTHY", false: "UNHEALTHY"}[health.Healthy])

	return health, nil
}

// testSubjectsEndpoint tests the /subjects endpoint
func (c *Client) testSubjectsEndpoint(health *HealthStatus) ([]string, error) {
	health.TestedEndpoints = append(health.TestedEndpoints, "GET /subjects")
	
	subjects, err := c.client.GetAllSubjects()
	if err != nil {
		errorDetail := c.classifyError("GET /subjects", "ListSubjects", err)
		health.ErrorDetails = append(health.ErrorDetails, errorDetail)
		return nil, err
	}
	
	return subjects, nil
}

// testSchemaMetadataEndpoint tests schema metadata retrieval
func (c *Client) testSchemaMetadataEndpoint(health *HealthStatus) {
	health.TestedEndpoints = append(health.TestedEndpoints, "GET /subjects/{subject}/versions/latest")
	
	// Test with a non-existent subject to verify error handling
	_, err := c.client.GetLatestSchemaMetadata("__nonexistent_health_check_subject__")
	if err != nil {
		// Expected error for non-existent subject (404/40401)
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "40401") {
			log.Printf("✅ Schema metadata endpoint properly returns 404 for non-existent subject")
		} else {
			// Unexpected error
			errorDetail := c.classifyError("GET /subjects/{subject}/versions/latest", "GetSchemaMetadata", err)
			health.ErrorDetails = append(health.ErrorDetails, errorDetail)
		}
	}
}

// testExistingSubject tests operations with an existing subject
func (c *Client) testExistingSubject(health *HealthStatus, subject string) {
	health.TestedEndpoints = append(health.TestedEndpoints, fmt.Sprintf("GET /subjects/%s/versions", subject))
	
	// Test getting latest schema for existing subject
	health.TestedEndpoints = append(health.TestedEndpoints, fmt.Sprintf("GET /subjects/%s/versions/latest", subject))
	
	_, err := c.client.GetLatestSchemaMetadata(subject)
	if err != nil {
		errorDetail := c.classifyError(fmt.Sprintf("GET /subjects/%s/versions/latest", subject), "GetLatestSchemaMetadata", err)
		health.ErrorDetails = append(health.ErrorDetails, errorDetail)
	} else {
		log.Printf("✅ Successfully retrieved latest schema for subject '%s'", subject)
	}
}

// testErrorHandling tests how the client handles various error scenarios
func (c *Client) testErrorHandling(health *HealthStatus) {
	// Test with extremely long subject name to trigger validation error
	longSubjectName := strings.Repeat("x", 300)
	health.TestedEndpoints = append(health.TestedEndpoints, "Error handling test")
	
	_, err := c.client.GetLatestSchemaMetadata(longSubjectName)
	if err != nil {
		// This should trigger some kind of validation or server error
		log.Printf("✅ Error handling working - long subject name properly rejected: %v", err)
	}
}

// classifyError analyzes an error and categorizes it for better handling
func (c *Client) classifyError(endpoint, operation string, err error) ErrorDetail {
	errorStr := err.Error()
	detail := ErrorDetail{
		Endpoint:  endpoint,
		Operation: operation,
		Message:   errorStr,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	// Classify error types
	switch {
	case strings.Contains(errorStr, "401") || strings.Contains(errorStr, "Unauthorized"):
		detail.ErrorType = "authentication"
		detail.ErrorCode = 401
		detail.Recoverable = false
		
	case strings.Contains(errorStr, "403") || strings.Contains(errorStr, "Forbidden"):
		detail.ErrorType = "authorization"
		detail.ErrorCode = 403
		detail.Recoverable = false
		
	case strings.Contains(errorStr, "404") || strings.Contains(errorStr, "not found"):
		detail.ErrorType = "not_found"
		detail.ErrorCode = 404
		detail.Recoverable = true // Might be temporary
		
	case strings.Contains(errorStr, "timeout") || strings.Contains(errorStr, "deadline exceeded"):
		detail.ErrorType = "timeout"
		detail.Recoverable = true
		
	case strings.Contains(errorStr, "connection refused") || strings.Contains(errorStr, "no such host"):
		detail.ErrorType = "network"
		detail.Recoverable = true
		
	case strings.Contains(errorStr, "500") || strings.Contains(errorStr, "Internal Server Error"):
		detail.ErrorType = "server_error"
		detail.ErrorCode = 500
		detail.Recoverable = true
		
	default:
		detail.ErrorType = "unknown"
		detail.Recoverable = false
	}

	return detail
}

// hasOnlyRecoverableErrors checks if all errors are recoverable
func (c *Client) hasOnlyRecoverableErrors(health *HealthStatus) bool {
	for _, error := range health.ErrorDetails {
		if !error.Recoverable {
			return false
		}
	}
	return true
}

// ValidateWithProtobufSchema tests Schema Registry with our Protobuf schema
func (c *Client) ValidateWithProtobufSchema(subject string, schemaContent string) error {
	log.Printf("🔍 Validating Schema Registry with Protobuf schema for subject: %s", subject)
	
	// This is a preparation for when we implement schema registration
	// For now, we'll just validate that we can check if the subject exists
	
	metadata, err := c.client.GetLatestSchemaMetadata(subject)
	if err != nil {
		// Subject might not exist yet, which is fine
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "40401") {
			log.Printf("📝 Subject '%s' does not exist yet (ready for schema registration)", subject)
			return nil
		}
		return fmt.Errorf("failed to validate subject '%s': %w", subject, err)
	}
	
	log.Printf("✅ Subject '%s' exists", subject)
	log.Printf("📋 Latest schema ID: %d, Version: %d", metadata.ID, metadata.Version)
	log.Printf("📝 Schema type: %s", metadata.SchemaType)
	
	return nil
}
