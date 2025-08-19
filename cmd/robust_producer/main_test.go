package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/go-sarama-sr/producer/internal/config"
	"github.com/go-sarama-sr/producer/pkg/errorhandling"
	"github.com/go-sarama-sr/producer/pkg/logging"
	"github.com/go-sarama-sr/producer/pkg/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAsyncProducer simulates Sarama async producer for testing
type MockAsyncProducer struct {
	mock.Mock
	inputChan   chan *sarama.ProducerMessage
	successChan chan *sarama.ProducerMessage
	errorChan   chan *sarama.ProducerError
	closed      bool
}

func NewMockAsyncProducer() *MockAsyncProducer {
	return &MockAsyncProducer{
		inputChan:   make(chan *sarama.ProducerMessage, 100),
		successChan: make(chan *sarama.ProducerMessage, 100),
		errorChan:   make(chan *sarama.ProducerError, 100),
	}
}

func (m *MockAsyncProducer) Input() chan<- *sarama.ProducerMessage {
	return m.inputChan
}

func (m *MockAsyncProducer) Successes() <-chan *sarama.ProducerMessage {
	return m.successChan
}

func (m *MockAsyncProducer) Errors() <-chan *sarama.ProducerError {
	return m.errorChan
}

func (m *MockAsyncProducer) AsyncClose() {
	m.closed = true
	close(m.inputChan)
	close(m.successChan)
	close(m.errorChan)
}

func (m *MockAsyncProducer) Close() error {
	m.AsyncClose()
	return nil
}

func (m *MockAsyncProducer) GetMetadata() (*sarama.MetadataResponse, error) {
	return nil, nil
}

func (m *MockAsyncProducer) IsTransactional() bool {
	return false
}

func (m *MockAsyncProducer) TxnStatus() sarama.ProducerTxnStatusFlag {
	return 0
}

func (m *MockAsyncProducer) BeginTxn() error {
	return nil
}

func (m *MockAsyncProducer) CommitTxn() error {
	return nil
}

func (m *MockAsyncProducer) AbortTxn() error {
	return nil
}

func (m *MockAsyncProducer) AddOffsetsToTxn(offsets map[string][]*sarama.PartitionOffsetMetadata, groupId string) error {
	return nil
}

func (m *MockAsyncProducer) AddMessageToTxn(msg *sarama.ConsumerMessage, groupId string, metadata *string) error {
	return nil
}

// SimulateSuccess simulates a successful message delivery
func (m *MockAsyncProducer) SimulateSuccess(msg *sarama.ProducerMessage) {
	msg.Partition = 0
	msg.Offset = int64(time.Now().UnixNano())
	m.successChan <- msg
}

// SimulateError simulates a message delivery error
func (m *MockAsyncProducer) SimulateError(msg *sarama.ProducerMessage, err error) {
	m.errorChan <- &sarama.ProducerError{
		Msg: msg,
		Err: err,
	}
}

// TestRobustProducerErrorClassification tests error classification logic
func TestRobustProducerErrorClassification(t *testing.T) {
	tests := []struct {
		name          string
		error         error
		expectedType  errorhandling.ErrorType
		expectedRetry bool
	}{
		{
			name:          "Network connection error",
			error:         errors.New("connection refused"),
			expectedType:  errorhandling.ErrorTypeTransient,
			expectedRetry: true,
		},
		{
			name:          "Authentication error",
			error:         errors.New("auth failed"),
			expectedType:  errorhandling.ErrorTypePermanent,
			expectedRetry: false,
		},
		{
			name:          "Broker unavailable",
			error:         errors.New("broker not available"),
			expectedType:  errorhandling.ErrorTypeTransient,
			expectedRetry: true,
		},
		{
			name:          "Message too large",
			error:         errors.New("message too large"),
			expectedType:  errorhandling.ErrorTypePermanent,
			expectedRetry: false,
		},
		{
			name:          "Quota exceeded",
			error:         errors.New("quota exceeded"),
			expectedType:  errorhandling.ErrorTypeTransient,
			expectedRetry: true,
		},
		{
			name:          "Serialization error",
			error:         errors.New("serialization failed"),
			expectedType:  errorhandling.ErrorTypePermanent,
			expectedRetry: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			classification := errorhandling.ClassifyError(tt.error)
			assert.Equal(t, tt.expectedType, classification.Type)
			assert.Equal(t, tt.expectedRetry, classification.Retryable)
		})
	}
}

// TestRetryPolicyBackoff tests exponential backoff calculation
func TestRetryPolicyBackoff(t *testing.T) {
	policy := errorhandling.DefaultRetryPolicy()

	tests := []struct {
		attempt  int
		expected time.Duration
	}{
		{0, 1 * time.Second},
		{1, 2 * time.Second},
		{2, 4 * time.Second},
		{3, 8 * time.Second},
		{4, 16 * time.Second},
		{5, 30 * time.Second}, // Capped at MaxDelay
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("attempt_%d", tt.attempt), func(t *testing.T) {
			backoff := policy.CalculateBackoff(tt.attempt)
			assert.Equal(t, tt.expected, backoff)
		})
	}
}

// TestMessageFailureTracking tests message failure record management
func TestMessageFailureTracking(t *testing.T) {
	msg := &sarama.ProducerMessage{
		Topic: "test-topic",
		Key:   sarama.StringEncoder("test-key"),
		Value: sarama.ByteEncoder("test-value"),
	}

	err := errors.New("network timeout")
	failure := errorhandling.NewMessageFailure(msg, err)

	assert.Equal(t, msg, failure.Message)
	assert.Equal(t, err, failure.Error)
	assert.Equal(t, 1, failure.AttemptCount)
	assert.True(t, failure.Classification.Retryable)
	assert.Equal(t, "network_error", failure.Classification.Code)

	// Test retry update
	policy := errorhandling.DefaultRetryPolicy()
	oldNextRetry := failure.NextRetry
	failure.UpdateForRetry(policy)

	assert.Equal(t, 2, failure.AttemptCount)
	assert.True(t, failure.NextRetry.After(oldNextRetry))
}

// TestRobustProducerStats tests statistics tracking
func TestRobustProducerStats(t *testing.T) {
	stats := &RobustProducerStats{
		StartTime: time.Now(),
	}

	// Test increment operations
	stats.IncrementProduced()
	stats.IncrementRetried()
	stats.IncrementDeadLettered()
	stats.IncrementTransientErrors()
	stats.IncrementPermanentErrors()
	stats.AddLatency(150.0)
	stats.AddLatency(250.0)

	produced, retried, dlq, transient, permanent, avgLatency, _ := stats.GetStats()

	assert.Equal(t, int64(1), produced)
	assert.Equal(t, int64(1), retried)
	assert.Equal(t, int64(1), dlq)
	assert.Equal(t, int64(1), transient)
	assert.Equal(t, int64(1), permanent)
	assert.Equal(t, 200.0, avgLatency) // (150 + 250) / 2
}

// TestRobustProducerEventHandling tests the event handling goroutines
func TestRobustProducerEventHandling(t *testing.T) {
	// Create test configuration
	cfg := &config.Config{
		Kafka: config.KafkaConfig{
			Topic: "test-topic",
		},
		App: config.AppConfig{
			LogLevel: "debug",
		},
	}

	// Create test logger
	logger, err := logging.NewProducerLogger("debug")
	assert.NoError(t, err)
	defer logger.Close()

	// Create mock producer
	mockProducer := NewMockAsyncProducer()
	mockDLQProducer := NewMockAsyncProducer()

	// Create test producer metrics
	metrics := metrics.NewProducerMetrics()

	// Create test robust producer
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rp := &RobustProducer{
		cfg:            cfg,
		producer:       mockProducer,
		dlqProducer:    mockDLQProducer,
		logger:         logger,
		metrics:        metrics,
		stats:          &RobustProducerStats{StartTime: time.Now()},
		retryPolicy:    errorhandling.DefaultRetryPolicy(),
		failedMessages: make(map[string]*errorhandling.MessageFailure),
		ctx:            ctx,
		cancel:         cancel,
		successEvents:  make(chan *sarama.ProducerMessage, 10),
		errorEvents:    make(chan *sarama.ProducerError, 10),
		retryQueue:     make(chan *errorhandling.MessageFailure, 10),
	}

	// Test successful message handling
	t.Run("SuccessfulMessageHandling", func(t *testing.T) {
		msg := &sarama.ProducerMessage{
			Topic:     "test-topic",
			Key:       sarama.StringEncoder("test-key"),
			Value:     sarama.ByteEncoder("test-value"),
			Timestamp: time.Now(),
			Metadata:  time.Now(),
		}

		rp.handleSuccessfulMessage(msg)

		produced, _, _, _, _, _, _ := rp.stats.GetStats()
		assert.Equal(t, int64(1), produced)
	})

	// Test error message handling
	t.Run("ErrorMessageHandling", func(t *testing.T) {
		msg := &sarama.ProducerMessage{
			Topic: "test-topic",
			Key:   sarama.StringEncoder("error-key"),
			Value: sarama.ByteEncoder("error-value"),
		}

		errEvent := &sarama.ProducerError{
			Msg: msg,
			Err: errors.New("network timeout"),
		}

		rp.handleErroredMessage(errEvent)

		// Check if message was added to failed messages
		key := string(msg.Key.(sarama.StringEncoder))
		_, exists := rp.failedMessages[key]
		assert.True(t, exists)

		// Check stats
		_, _, _, transient, _, _, _ := rp.stats.GetStats()
		assert.Equal(t, int64(1), transient)
	})

	// Test DLQ handling
	t.Run("DeadLetterQueueHandling", func(t *testing.T) {
		msg := &sarama.ProducerMessage{
			Topic: "test-topic",
			Key:   sarama.StringEncoder("dlq-key"),
			Value: sarama.ByteEncoder("dlq-value"),
		}

		failure := errorhandling.NewMessageFailure(msg, errors.New("auth failed"))
		rp.sendToDeadLetterQueue(failure, "non_retryable_error")

		// Check if DLQ message was sent (mock would receive it)
		select {
		case dlqMsg := <-mockDLQProducer.inputChan:
			assert.Equal(t, "test-topic.dlq", dlqMsg.Topic)
			assert.Equal(t, msg.Key, dlqMsg.Key)
			assert.Equal(t, msg.Value, dlqMsg.Value)

			// Check headers
			foundReason := false
			for _, header := range dlqMsg.Headers {
				if string(header.Key) == "failure-reason" {
					assert.Equal(t, "non_retryable_error", string(header.Value))
					foundReason = true
				}
			}
			assert.True(t, foundReason)
		case <-time.After(1 * time.Second):
			t.Fatal("DLQ message not sent within timeout")
		}

		// Check stats
		_, _, dlq, _, _, _, _ := rp.stats.GetStats()
		assert.Equal(t, int64(1), dlq)
	})
}

// BenchmarkErrorClassification benchmarks error classification performance
func BenchmarkErrorClassification(b *testing.B) {
	errors := []error{
		fmt.Errorf("connection refused"),
		fmt.Errorf("auth failed"),
		fmt.Errorf("broker not available"),
		fmt.Errorf("message too large"),
		fmt.Errorf("quota exceeded"),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := errors[i%len(errors)]
		errorhandling.ClassifyError(err)
	}
}

// TestMain sets up test environment
func TestMain(m *testing.M) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	m.Run()
}
