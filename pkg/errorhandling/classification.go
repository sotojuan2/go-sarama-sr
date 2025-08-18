package errorhandling

import (
	"fmt"
	"strings"
	"time"

	"github.com/IBM/sarama"
)

// ErrorType categorizes different types of errors
type ErrorType string

const (
	ErrorTypeTransient    ErrorType = "transient"
	ErrorTypePermanent    ErrorType = "permanent"
	ErrorTypeRetryable    ErrorType = "retryable"
	ErrorTypeNonRetryable ErrorType = "non_retryable"
)

// ErrorClassification contains error analysis results
type ErrorClassification struct {
	Type        ErrorType
	Code        string
	Retryable   bool
	MaxRetries  int
	BackoffMs   int
	Description string
}

// ClassifyError analyzes a Sarama producer error and determines retry strategy
func ClassifyError(err error) ErrorClassification {
	if err == nil {
		return ErrorClassification{
			Type:        ErrorTypePermanent,
			Code:        "unknown",
			Retryable:   false,
			Description: "Unknown error",
		}
	}

	errStr := strings.ToLower(err.Error())

	// Network-related errors (usually transient)
	if strings.Contains(errStr, "connection") ||
		strings.Contains(errStr, "network") ||
		strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "dial") ||
		strings.Contains(errStr, "broken pipe") {
		return ErrorClassification{
			Type:        ErrorTypeTransient,
			Code:        "network_error",
			Retryable:   true,
			MaxRetries:  5,
			BackoffMs:   1000,
			Description: "Network connectivity issue",
		}
	}

	// Kafka broker errors
	if strings.Contains(errStr, "broker") {
		if strings.Contains(errStr, "not available") ||
			strings.Contains(errStr, "leader not available") {
			return ErrorClassification{
				Type:        ErrorTypeTransient,
				Code:        "broker_unavailable",
				Retryable:   true,
				MaxRetries:  3,
				BackoffMs:   2000,
				Description: "Kafka broker temporarily unavailable",
			}
		}
	}

	// Authentication/Authorization errors (permanent)
	if strings.Contains(errStr, "auth") ||
		strings.Contains(errStr, "sasl") ||
		strings.Contains(errStr, "unauthorized") ||
		strings.Contains(errStr, "forbidden") {
		return ErrorClassification{
			Type:        ErrorTypePermanent,
			Code:        "auth_error",
			Retryable:   false,
			Description: "Authentication or authorization failure",
		}
	}

	// Topic/Partition errors
	if strings.Contains(errStr, "topic") {
		if strings.Contains(errStr, "not exist") ||
			strings.Contains(errStr, "unknown topic") {
			return ErrorClassification{
				Type:        ErrorTypePermanent,
				Code:        "topic_not_found",
				Retryable:   false,
				Description: "Topic does not exist",
			}
		}
	}

	// Message size errors (permanent)
	if strings.Contains(errStr, "message too large") ||
		strings.Contains(errStr, "record batch too large") {
		return ErrorClassification{
			Type:        ErrorTypePermanent,
			Code:        "message_too_large",
			Retryable:   false,
			Description: "Message exceeds maximum allowed size",
		}
	}

	// Quota exceeded (transient)
	if strings.Contains(errStr, "quota") ||
		strings.Contains(errStr, "throttle") {
		return ErrorClassification{
			Type:        ErrorTypeTransient,
			Code:        "quota_exceeded",
			Retryable:   true,
			MaxRetries:  3,
			BackoffMs:   5000,
			Description: "Rate limit or quota exceeded",
		}
	}

	// Serialization errors (permanent)
	if strings.Contains(errStr, "serializ") ||
		strings.Contains(errStr, "schema") ||
		strings.Contains(errStr, "encoding") {
		return ErrorClassification{
			Type:        ErrorTypePermanent,
			Code:        "serialization_error",
			Retryable:   false,
			Description: "Message serialization or schema error",
		}
	}

	// Default: treat as retryable transient error
	return ErrorClassification{
		Type:        ErrorTypeTransient,
		Code:        "unknown_transient",
		Retryable:   true,
		MaxRetries:  2,
		BackoffMs:   3000,
		Description: "Unknown transient error",
	}
}

// RetryPolicy defines retry behavior for different error types
type RetryPolicy struct {
	MaxRetries    int
	InitialDelay  time.Duration
	MaxDelay      time.Duration
	BackoffFactor float64
}

// DefaultRetryPolicy returns the default retry policy
func DefaultRetryPolicy() RetryPolicy {
	return RetryPolicy{
		MaxRetries:    3,
		InitialDelay:  1 * time.Second,
		MaxDelay:      30 * time.Second,
		BackoffFactor: 2.0,
	}
}

// CalculateBackoff calculates the delay for a retry attempt
func (rp RetryPolicy) CalculateBackoff(attempt int) time.Duration {
	if attempt <= 0 {
		return rp.InitialDelay
	}

	delay := rp.InitialDelay
	for i := 0; i < attempt; i++ {
		delay = time.Duration(float64(delay) * rp.BackoffFactor)
		if delay > rp.MaxDelay {
			delay = rp.MaxDelay
			break
		}
	}

	return delay
}

// ShouldRetry determines if an error should be retried
func (rp RetryPolicy) ShouldRetry(classification ErrorClassification, attempt int) bool {
	if !classification.Retryable {
		return false
	}

	maxRetries := rp.MaxRetries
	if classification.MaxRetries > 0 {
		maxRetries = classification.MaxRetries
	}

	return attempt < maxRetries
}

// MessageFailure represents a failed message with retry information
type MessageFailure struct {
	Message        *sarama.ProducerMessage
	Error          error
	Classification ErrorClassification
	AttemptCount   int
	FirstAttempt   time.Time
	LastAttempt    time.Time
	NextRetry      time.Time
}

// NewMessageFailure creates a new message failure record
func NewMessageFailure(message *sarama.ProducerMessage, err error) *MessageFailure {
	classification := ClassifyError(err)
	now := time.Now()

	return &MessageFailure{
		Message:        message,
		Error:          err,
		Classification: classification,
		AttemptCount:   1,
		FirstAttempt:   now,
		LastAttempt:    now,
		NextRetry:      now.Add(time.Duration(classification.BackoffMs) * time.Millisecond),
	}
}

// ShouldRetryNow checks if the message should be retried now
func (mf *MessageFailure) ShouldRetryNow(policy RetryPolicy) bool {
	if !policy.ShouldRetry(mf.Classification, mf.AttemptCount) {
		return false
	}
	return time.Now().After(mf.NextRetry)
}

// UpdateForRetry updates the failure record for a retry attempt
func (mf *MessageFailure) UpdateForRetry(policy RetryPolicy) {
	mf.AttemptCount++
	mf.LastAttempt = time.Now()
	mf.NextRetry = mf.LastAttempt.Add(policy.CalculateBackoff(mf.AttemptCount))
}

// GetMessageID extracts or generates a message ID for tracking
func GetMessageID(message *sarama.ProducerMessage) string {
	// Try to get ID from headers first
	for _, header := range message.Headers {
		if string(header.Key) == "message-id" || string(header.Key) == "shoe-id" {
			return string(header.Value)
		}
	}

	// Generate a fallback ID using timestamp and topic
	return fmt.Sprintf("%s-%d", message.Topic, time.Now().UnixNano())
}
