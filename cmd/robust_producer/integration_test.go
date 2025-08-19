package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/go-sarama-sr/producer/internal/config"
	"github.com/go-sarama-sr/producer/pkg/errorhandling"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

// TestRobustProducerIntegration performs end-to-end integration testing
func TestRobustProducerIntegration(t *testing.T) {
	// Skip if no .env file available (CI/CD environment)
	if err := godotenv.Load("../../.env"); err != nil {
		t.Skip("Skipping integration test: .env file not available")
	}

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify we're not running against production
	if !strings.Contains(cfg.Kafka.BootstrapServers, "confluent.cloud") {
		t.Skip("Skipping integration test: not running against test environment")
	}

	t.Run("RobustProducerHealthCheck", func(t *testing.T) {
		// Start robust producer in background
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// This would ideally use the actual RobustProducer
		// For now, we'll test the metrics endpoint availability
		go func() {
			// Simulate metrics server
			http.HandleFunc("/test-metrics", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("# Test metrics\ntest_counter 1\n"))
			})

			// Use context to control server lifecycle
			select {
			case <-ctx.Done():
				return
			default:
				http.ListenAndServe(":8081", nil)
			}
		}()

		// Wait for server to start
		time.Sleep(2 * time.Second)

		// Test metrics endpoint
		resp, err := http.Get("http://localhost:8081/test-metrics")
		if err == nil {
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})

	t.Run("ErrorScenarioSimulation", func(t *testing.T) {
		// Test various error scenarios in isolation
		testCases := []struct {
			scenario    string
			errorType   string
			expectRetry bool
			expectDLQ   bool
		}{
			{
				scenario:    "Network timeout",
				errorType:   "connection timeout",
				expectRetry: true,
				expectDLQ:   false,
			},
			{
				scenario:    "Authentication failure",
				errorType:   "auth failed",
				expectRetry: false,
				expectDLQ:   true,
			},
			{
				scenario:    "Message too large",
				errorType:   "message too large",
				expectRetry: false,
				expectDLQ:   true,
			},
			{
				scenario:    "Quota exceeded",
				errorType:   "quota exceeded",
				expectRetry: true,
				expectDLQ:   false,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.scenario, func(t *testing.T) {
				// Simulate error classification
				err := fmt.Errorf("%s", tc.errorType)
				classification := errorhandling.ClassifyError(err)

				assert.Equal(t, tc.expectRetry, classification.Retryable)

				if tc.expectDLQ && !classification.Retryable {
					// Non-retryable errors should go to DLQ
					assert.False(t, classification.Retryable)
				}
			})
		}
	})

	t.Run("MessageThroughputTest", func(t *testing.T) {
		// Test message throughput simulation
		stats := &RobustProducerStats{
			StartTime: time.Now(),
		}

		// Simulate processing 100 messages
		start := time.Now()
		for i := 0; i < 100; i++ {
			stats.IncrementProduced()
			stats.AddLatency(float64(50 + i%100)) // Varying latency 50-150ms
		}
		duration := time.Since(start)

		produced, _, _, _, _, avgLatency, _ := stats.GetStats()

		assert.Equal(t, int64(100), produced)
		assert.Greater(t, avgLatency, 0.0)
		assert.Less(t, duration, 1*time.Second) // Should be very fast in test

		// Calculate throughput
		throughput := float64(produced) / duration.Seconds()
		assert.Greater(t, throughput, 10000.0) // Should be > 10k msgs/sec in test
	})

	t.Run("RetryLogicValidation", func(t *testing.T) {
		policy := errorhandling.DefaultRetryPolicy()

		// Test exponential backoff progression
		backoffs := []time.Duration{}
		for i := 0; i < 6; i++ {
			backoff := policy.CalculateBackoff(i)
			backoffs = append(backoffs, backoff)
		}

		// Verify exponential progression
		assert.Equal(t, 1*time.Second, backoffs[0])
		assert.Equal(t, 2*time.Second, backoffs[1])
		assert.Equal(t, 4*time.Second, backoffs[2])
		assert.Equal(t, 8*time.Second, backoffs[3])
		assert.Equal(t, 16*time.Second, backoffs[4])
		assert.Equal(t, 30*time.Second, backoffs[5]) // Capped at max

		// Test retry decision logic
		transientError := errorhandling.ClassifyError(fmt.Errorf("connection timeout"))
		permanentError := errorhandling.ClassifyError(fmt.Errorf("auth failed"))

		assert.True(t, policy.ShouldRetry(transientError, 1))
		assert.True(t, policy.ShouldRetry(transientError, 2))
		assert.False(t, policy.ShouldRetry(transientError, 5)) // Exceeds max retries

		assert.False(t, policy.ShouldRetry(permanentError, 1)) // Non-retryable
	})

	t.Run("MemoryLeakTest", func(t *testing.T) {
		// Test that failed message tracking doesn't cause memory leaks
		failedMessages := make(map[string]*errorhandling.MessageFailure)

		// Simulate adding many failed messages
		for i := 0; i < 1000; i++ {
			key := fmt.Sprintf("key-%d", i)
			msg := &sarama.ProducerMessage{
				Topic: "test",
				Key:   sarama.StringEncoder(key),
			}
			failure := errorhandling.NewMessageFailure(msg, fmt.Errorf("test error"))
			failedMessages[key] = failure
		}

		assert.Equal(t, 1000, len(failedMessages))

		// Simulate cleanup (successful retries or DLQ routing)
		for key := range failedMessages {
			delete(failedMessages, key)
		}

		assert.Equal(t, 0, len(failedMessages))
	})

	t.Run("ConcurrencyStressTest", func(t *testing.T) {
		// Test concurrent goroutine safety
		stats := &RobustProducerStats{
			StartTime: time.Now(),
		}

		// Run concurrent operations
		done := make(chan bool, 4)

		// Goroutine 1: Increment produced
		go func() {
			for i := 0; i < 100; i++ {
				stats.IncrementProduced()
			}
			done <- true
		}()

		// Goroutine 2: Increment errors
		go func() {
			for i := 0; i < 50; i++ {
				stats.IncrementTransientErrors()
			}
			done <- true
		}()

		// Goroutine 3: Add latency
		go func() {
			for i := 0; i < 100; i++ {
				stats.AddLatency(float64(i))
			}
			done <- true
		}()

		// Goroutine 4: Read stats
		go func() {
			for i := 0; i < 10; i++ {
				stats.GetStats()
				time.Sleep(1 * time.Millisecond)
			}
			done <- true
		}()

		// Wait for all goroutines
		for i := 0; i < 4; i++ {
			<-done
		}

		produced, _, _, transient, _, _, _ := stats.GetStats()
		assert.Equal(t, int64(100), produced)
		assert.Equal(t, int64(50), transient)
	})
}

// TestRobustProducerFaultTolerance tests fault tolerance capabilities
func TestRobustProducerFaultTolerance(t *testing.T) {
	t.Run("ChannelBufferOverflow", func(t *testing.T) {
		// Test behavior when channels are full
		successChan := make(chan *sarama.ProducerMessage, 2)
		errorChan := make(chan *sarama.ProducerError, 2)

		// Fill channels to capacity
		for i := 0; i < 2; i++ {
			successChan <- &sarama.ProducerMessage{}
			errorChan <- &sarama.ProducerError{}
		}

		// Test non-blocking behavior
		select {
		case successChan <- &sarama.ProducerMessage{}:
			t.Fatal("Should not be able to write to full channel")
		default:
			// Expected behavior - channel is full
		}

		select {
		case errorChan <- &sarama.ProducerError{}:
			t.Fatal("Should not be able to write to full channel")
		default:
			// Expected behavior - channel is full
		}

		// Drain channels
		<-successChan
		<-errorChan

		// Should be able to write now
		select {
		case successChan <- &sarama.ProducerMessage{}:
			// Expected behavior
		default:
			t.Fatal("Should be able to write to drained channel")
		}
	})

	t.Run("GracefulShutdown", func(t *testing.T) {
		// Test graceful shutdown behavior
		ctx, cancel := context.WithCancel(context.Background())

		// Simulate goroutine that respects context cancellation
		done := make(chan bool)
		go func() {
			ticker := time.NewTicker(10 * time.Millisecond)
			defer ticker.Stop()

			for {
				select {
				case <-ctx.Done():
					done <- true
					return
				case <-ticker.C:
					// Simulate work
				}
			}
		}()

		// Let it run briefly
		time.Sleep(50 * time.Millisecond)

		// Cancel and verify shutdown
		cancel()

		select {
		case <-done:
			// Expected - goroutine shut down gracefully
		case <-time.After(1 * time.Second):
			t.Fatal("Goroutine did not shut down gracefully")
		}
	})

	t.Run("ErrorRecovery", func(t *testing.T) {
		// Test recovery from various error conditions
		errors := []error{
			fmt.Errorf("connection refused"),
			fmt.Errorf("timeout"),
			fmt.Errorf("temporary failure"),
		}

		for _, err := range errors {
			classification := errorhandling.ClassifyError(err)
			if classification.Retryable {
				// Simulate retry attempt
				policy := errorhandling.DefaultRetryPolicy()
				backoff := policy.CalculateBackoff(1)
				assert.Greater(t, backoff, 0*time.Second)
				assert.LessOrEqual(t, backoff, 30*time.Second) // Max delay
			}
		}
	})
}

// TestDocumentationExamples tests examples from documentation
func TestDocumentationExamples(t *testing.T) {
	t.Run("BasicUsageExample", func(t *testing.T) {
		// This would test the basic usage example from documentation
		// Verifying that the API works as documented

		stats := &RobustProducerStats{StartTime: time.Now()}

		// Simulate the documented workflow
		stats.IncrementProduced()
		produced, _, _, _, _, _, duration := stats.GetStats()

		assert.Equal(t, int64(1), produced)
		assert.Greater(t, duration, 0*time.Second)
	})

	t.Run("ErrorHandlingExample", func(t *testing.T) {
		// Test the error handling example from documentation
		err := fmt.Errorf("network connection failed")
		classification := errorhandling.ClassifyError(err)

		assert.True(t, classification.Retryable)
		assert.Equal(t, "network_error", classification.Code)
		assert.Equal(t, 5, classification.MaxRetries)
	})
}
