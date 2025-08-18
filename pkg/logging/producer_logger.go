package logging

import (
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger wraps zap logger with producer-specific functionality
type Logger struct {
	*zap.SugaredLogger
	structured *zap.Logger
}

// NewProducerLogger creates a new structured logger for the producer
func NewProducerLogger(level string) (*Logger, error) {
	// Configure log level
	zapLevel := zapcore.InfoLevel
	switch strings.ToLower(level) {
	case "debug":
		zapLevel = zapcore.DebugLevel
	case "info":
		zapLevel = zapcore.InfoLevel
	case "warn", "warning":
		zapLevel = zapcore.WarnLevel
	case "error":
		zapLevel = zapcore.ErrorLevel
	}

	// Configure encoder for structured output
	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(zapLevel)
	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}

	// Use JSON encoding for structured logs in production
	config.Encoding = "console" // Change to "json" for production
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.LevelKey = "level"
	config.EncoderConfig.MessageKey = "message"
	config.EncoderConfig.CallerKey = "caller"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	// Build logger
	logger, err := config.Build(zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	if err != nil {
		return nil, err
	}

	return &Logger{
		SugaredLogger: logger.Sugar(),
		structured:    logger,
	}, nil
}

// LogMessageSuccess logs a successful message production with structured fields
func (l *Logger) LogMessageSuccess(messageID, topic string, partition int32, offset int64, latencyMs float64) {
	l.structured.Info("Message produced successfully",
		zap.String("event_type", "message_success"),
		zap.String("message_id", messageID),
		zap.String("topic", topic),
		zap.Int32("partition", partition),
		zap.Int64("offset", offset),
		zap.Float64("latency_ms", latencyMs),
	)
}

// LogMessageError logs a message production error with structured fields
func (l *Logger) LogMessageError(messageID, topic, errorType, errorCode string, err error, retryCount int) {
	l.structured.Error("Message production failed",
		zap.String("event_type", "message_error"),
		zap.String("message_id", messageID),
		zap.String("topic", topic),
		zap.String("error_type", errorType),
		zap.String("error_code", errorCode),
		zap.Error(err),
		zap.Int("retry_count", retryCount),
	)
}

// LogEventProcessing logs event processing information
func (l *Logger) LogEventProcessing(eventType string, processingTimeMs float64, queueSize int) {
	l.structured.Debug("Event processed",
		zap.String("event_type", eventType),
		zap.Float64("processing_time_ms", processingTimeMs),
		zap.Int("queue_size", queueSize),
	)
}

// LogProducerHealth logs producer health information
func (l *Logger) LogProducerHealth(totalProduced, totalErrors int64, successRate float64, avgLatencyMs float64) {
	l.structured.Info("Producer health status",
		zap.String("event_type", "health_status"),
		zap.Int64("total_produced", totalProduced),
		zap.Int64("total_errors", totalErrors),
		zap.Float64("success_rate", successRate),
		zap.Float64("avg_latency_ms", avgLatencyMs),
	)
}

// LogDeadLetterQueue logs when a message is sent to dead letter queue
func (l *Logger) LogDeadLetterQueue(messageID, originalTopic, dlqTopic, reason string, attemptCount int) {
	l.structured.Warn("Message sent to dead letter queue",
		zap.String("event_type", "dead_letter_queue"),
		zap.String("message_id", messageID),
		zap.String("original_topic", originalTopic),
		zap.String("dlq_topic", dlqTopic),
		zap.String("reason", reason),
		zap.Int("attempt_count", attemptCount),
	)
}

// LogGoroutineStarted logs when a goroutine starts
func (l *Logger) LogGoroutineStarted(goroutineType string) {
	l.structured.Debug("Goroutine started",
		zap.String("event_type", "goroutine_lifecycle"),
		zap.String("goroutine_type", goroutineType),
		zap.String("action", "started"),
	)
}

// LogGoroutineStopped logs when a goroutine stops
func (l *Logger) LogGoroutineStopped(goroutineType string) {
	l.structured.Debug("Goroutine stopped",
		zap.String("event_type", "goroutine_lifecycle"),
		zap.String("goroutine_type", goroutineType),
		zap.String("action", "stopped"),
	)
}

// Sync flushes any buffered log entries
func (l *Logger) Sync() error {
	return l.structured.Sync()
}

// Close closes the logger
func (l *Logger) Close() error {
	return l.Sync()
}
