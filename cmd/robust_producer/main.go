package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"net/http"

	"github.com/IBM/sarama"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde/protobuf"
	"github.com/go-sarama-sr/producer/internal/config"
	"github.com/go-sarama-sr/producer/pkg/errorhandling"
	"github.com/go-sarama-sr/producer/pkg/generator"
	"github.com/go-sarama-sr/producer/pkg/logging"
	"github.com/go-sarama-sr/producer/pkg/metrics"
	srClient "github.com/go-sarama-sr/producer/pkg/schemaregistry"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// RobustProducerStats tracks comprehensive production metrics
type RobustProducerStats struct {
	mu                   sync.RWMutex
	MessagesProduced     int64
	MessagesRetried      int64
	MessagesDeadLettered int64
	ErrorsTransient      int64
	ErrorsPermanent      int64
	StartTime            time.Time
	LastMessageTime      time.Time
	TotalLatencyMs       float64
	MessageCount         int64
}

func (rps *RobustProducerStats) IncrementProduced() {
	rps.mu.Lock()
	defer rps.mu.Unlock()
	rps.MessagesProduced++
	rps.LastMessageTime = time.Now()
}

func (rps *RobustProducerStats) IncrementRetried() {
	rps.mu.Lock()
	defer rps.mu.Unlock()
	rps.MessagesRetried++
}

func (rps *RobustProducerStats) IncrementDeadLettered() {
	rps.mu.Lock()
	defer rps.mu.Unlock()
	rps.MessagesDeadLettered++
}

func (rps *RobustProducerStats) IncrementTransientErrors() {
	rps.mu.Lock()
	defer rps.mu.Unlock()
	rps.ErrorsTransient++
}

func (rps *RobustProducerStats) IncrementPermanentErrors() {
	rps.mu.Lock()
	defer rps.mu.Unlock()
	rps.ErrorsPermanent++
}

func (rps *RobustProducerStats) AddLatency(latencyMs float64) {
	rps.mu.Lock()
	defer rps.mu.Unlock()
	rps.TotalLatencyMs += latencyMs
	rps.MessageCount++
}

func (rps *RobustProducerStats) GetStats() (int64, int64, int64, int64, int64, float64, time.Duration) {
	rps.mu.RLock()
	defer rps.mu.RUnlock()

	avgLatency := float64(0)
	if rps.MessageCount > 0 {
		avgLatency = rps.TotalLatencyMs / float64(rps.MessageCount)
	}

	duration := time.Since(rps.StartTime)
	return rps.MessagesProduced, rps.MessagesRetried, rps.MessagesDeadLettered,
		rps.ErrorsTransient, rps.ErrorsPermanent, avgLatency, duration
}

// RobustProducer manages robust asynchronous message production
type RobustProducer struct {
	cfg            *config.Config
	srClient       *srClient.Client
	generator      *generator.ShoeGenerator
	serializer     *protobuf.Serializer
	producer       sarama.AsyncProducer
	dlqProducer    sarama.AsyncProducer
	logger         *logging.Logger
	metrics        *metrics.ProducerMetrics
	stats          *RobustProducerStats
	retryPolicy    errorhandling.RetryPolicy
	failedMessages map[string]*errorhandling.MessageFailure
	failedMsgMutex sync.RWMutex
	ctx            context.Context
	cancel         context.CancelFunc
	wg             sync.WaitGroup

	// Buffered channels for event processing
	successEvents chan *sarama.ProducerMessage
	errorEvents   chan *sarama.ProducerError
	retryQueue    chan *errorhandling.MessageFailure
}

func main() {
	fmt.Println("🚀 Starting Robust Asynchronous Kafka Producer...")

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		fmt.Printf("Warning: Could not load .env file: %v\n", err)
	}

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("❌ Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logger, err := logging.NewProducerLogger(cfg.App.LogLevel)
	if err != nil {
		fmt.Printf("❌ Error creating logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Close()

	logger.Info("🎯 Robust Producer Configuration:",
		"kafka_cluster", cfg.Kafka.BootstrapServers,
		"topic", cfg.Kafka.Topic,
		"message_interval", cfg.App.MessageInterval,
		"schema_registry", cfg.SchemaRegistry.URL,
		"log_level", cfg.App.LogLevel)

	// Start metrics server
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		logger.Info("📊 Metrics server starting on :8080/metrics")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			logger.Error("Failed to start metrics server", "error", err)
		}
	}()

	// Create robust producer
	robustProducer, err := NewRobustProducer(cfg, logger)
	if err != nil {
		logger.Fatal("❌ Failed to create robust producer", "error", err)
	}
	defer robustProducer.Close()

	// Set up graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt signals
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// Start the robust production system
	logger.Info("🔄 Starting robust production system...")
	go robustProducer.Start(ctx)

	// Wait for shutdown signal
	<-signalChan
	logger.Info("⏹️ Shutdown signal received. Stopping production...")

	// Cancel context and wait for graceful shutdown
	cancel()
	robustProducer.WaitForShutdown()

	// Print final comprehensive stats
	robustProducer.PrintFinalStats()
	logger.Info("✅ Robust producer shutdown completed")
}

// NewRobustProducer creates a new robust producer instance
func NewRobustProducer(cfg *config.Config, logger *logging.Logger) (*RobustProducer, error) {
	logger.Info("🔧 Initializing robust producer components...")

	// Create Schema Registry client
	srClientWrapper, err := srClient.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Schema Registry client: %w", err)
	}

	// Verify Schema Registry connectivity
	logger.Info("🔍 Testing Schema Registry connectivity...")
	healthResult, err := srClientWrapper.ComprehensiveHealthCheck()
	if err != nil {
		return nil, fmt.Errorf("Schema Registry health check failed: %w", err)
	}
	logger.Info("✅ Schema Registry connectivity verified", "status", healthResult)

	// Initialize components
	shoeGen := generator.NewShoeGenerator()
	logger.Info("✅ Random shoe generator initialized")

	serializer, err := protobuf.NewSerializer(srClientWrapper.GetClient(), serde.ValueSerde, protobuf.NewSerializerConfig())
	if err != nil {
		return nil, fmt.Errorf("failed to create protobuf serializer: %w", err)
	}
	logger.Info("✅ Protobuf serializer created")

	// Create main producer
	producer, err := createAsyncProducer(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create main producer: %w", err)
	}
	logger.Info("✅ Main Kafka async producer created")

	// Create DLQ producer
	dlqProducer, err := createAsyncProducer(cfg)
	if err != nil {
		producer.Close()
		return nil, fmt.Errorf("failed to create DLQ producer: %w", err)
	}
	logger.Info("✅ Dead Letter Queue producer created")

	// Initialize metrics
	producerMetrics := metrics.NewProducerMetrics()
	logger.Info("✅ Prometheus metrics initialized")

	ctx, cancel := context.WithCancel(context.Background())

	return &RobustProducer{
		cfg:            cfg,
		srClient:       srClientWrapper,
		generator:      shoeGen,
		serializer:     serializer,
		producer:       producer,
		dlqProducer:    dlqProducer,
		logger:         logger,
		metrics:        producerMetrics,
		stats:          &RobustProducerStats{StartTime: time.Now()},
		retryPolicy:    errorhandling.DefaultRetryPolicy(),
		failedMessages: make(map[string]*errorhandling.MessageFailure),
		ctx:            ctx,
		cancel:         cancel,

		// Buffered channels to prevent blocking
		successEvents: make(chan *sarama.ProducerMessage, 1000),
		errorEvents:   make(chan *sarama.ProducerError, 1000),
		retryQueue:    make(chan *errorhandling.MessageFailure, 500),
	}, nil
}

// createAsyncProducer creates a configured async producer
func createAsyncProducer(cfg *config.Config) (sarama.AsyncProducer, error) {
	saramaConfig := sarama.NewConfig()
	saramaConfig.Producer.RequiredAcks = sarama.WaitForAll
	saramaConfig.Producer.Retry.Max = cfg.Kafka.RetryMax
	saramaConfig.Producer.Return.Successes = true
	saramaConfig.Producer.Return.Errors = true
	saramaConfig.Producer.Timeout = cfg.Kafka.Timeout

	// SASL configuration for Confluent Cloud
	saramaConfig.Net.SASL.Enable = true
	saramaConfig.Net.SASL.Mechanism = sarama.SASLTypePlaintext
	saramaConfig.Net.SASL.User = cfg.Kafka.APIKey
	saramaConfig.Net.SASL.Password = cfg.Kafka.APISecret
	saramaConfig.Net.TLS.Enable = true

	// Buffering for performance
	saramaConfig.Producer.Flush.Frequency = 100 * time.Millisecond
	saramaConfig.Producer.Flush.Messages = 100

	brokers := strings.Split(cfg.Kafka.BootstrapServers, ",")
	return sarama.NewAsyncProducer(brokers, saramaConfig)
}

// Start begins the robust production system
func (rp *RobustProducer) Start(ctx context.Context) {
	rp.logger.Info("🚀 Starting robust production system goroutines...")

	// Update metrics for active goroutines
	rp.metrics.SetActiveGoroutines(6) // Will start 6 goroutines

	// Start event processing goroutines
	rp.wg.Add(6)
	go rp.handleSuccessEvents(ctx)
	go rp.handleErrorEvents(ctx)
	go rp.processRetryQueue(ctx)
	go rp.drainProducerChannels(ctx)
	go rp.startStatsReporting(ctx)
	go rp.startProduction(ctx)

	rp.logger.Info("✅ All robust production goroutines started")
}

// startProduction generates and sends messages continuously
func (rp *RobustProducer) startProduction(ctx context.Context) {
	defer rp.wg.Done()
	rp.logger.Info("📤 Starting continuous message production...")

	ticker := time.NewTicker(rp.cfg.App.MessageInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			rp.logger.Info("⏹️ Production stopped by context cancellation")
			return
		case <-ticker.C:
			if err := rp.produceMessage(ctx); err != nil {
				rp.logger.Error("Failed to produce message", "error", err)
				rp.metrics.RecordMessageError(rp.cfg.Kafka.Topic, "production", "failure")
			}
		}
	}
}

// produceMessage creates and sends a single message
func (rp *RobustProducer) produceMessage(ctx context.Context) error {
	startTime := time.Now()

		// Generate shoe data
	shoe := rp.generator.GenerateRandomShoe()
	
	// Serialize message
	payload, err := rp.serializer.Serialize(rp.cfg.Kafka.Topic, shoe)
	if err != nil {
		rp.logger.Error("Failed to serialize message",
			"error", err,
			"shoe_id", shoe.Id)
		rp.metrics.RecordMessageError(rp.cfg.Kafka.Topic, "serialization", "error")
		return fmt.Errorf("serialization failed: %w", err)
	}

	// Create message
	message := &sarama.ProducerMessage{
		Topic:     rp.cfg.Kafka.Topic,
		Key:       sarama.StringEncoder(fmt.Sprintf("%d", shoe.Id)),
		Value:     sarama.ByteEncoder(payload),
		Timestamp: time.Now(),
		Metadata:  startTime, // Store start time for latency calculation
	}

	// Send to producer
	select {
	case rp.producer.Input() <- message:
		rp.logger.Debug("Message queued for production",
			"shoe_id", shoe.Id,
			"brand", shoe.Brand,
			"name", shoe.Name)
		rp.metrics.RecordMessageProduced(rp.cfg.Kafka.Topic, 0) // Will be updated when partition is known
		return nil
	case <-ctx.Done():
		return fmt.Errorf("production cancelled")
	default:
		// Producer input channel is full
		rp.logger.Warn("Producer input channel full, message dropped")
		rp.metrics.RecordMessageError(rp.cfg.Kafka.Topic, "channel", "full")
		return fmt.Errorf("producer channel full")
	}
}

// handleSuccessEvents processes successful message deliveries
func (rp *RobustProducer) handleSuccessEvents(ctx context.Context) {
	defer rp.wg.Done()
	rp.logger.Info("✅ Starting success event handler...")

	for {
		select {
		case <-ctx.Done():
			rp.logger.Info("⏹️ Success handler stopped")
			return
		case success := <-rp.successEvents:
			rp.handleSuccessfulMessage(success)
		}
	}
}

// handleSuccessfulMessage processes a successful message delivery
func (rp *RobustProducer) handleSuccessfulMessage(msg *sarama.ProducerMessage) {
	// Calculate latency
	startTime, ok := msg.Metadata.(time.Time)
	var latencyMs float64
	if ok {
		latencyMs = float64(time.Since(startTime).Nanoseconds()) / 1e6
		rp.metrics.RecordProductionLatency(msg.Topic, latencyMs/1000) // Convert to seconds for Prometheus
		rp.stats.AddLatency(latencyMs)
	}

	// Update statistics
	rp.stats.IncrementProduced()
	rp.metrics.RecordMessageProduced(msg.Topic, msg.Partition)

	// Log success with structured fields
	key := string(msg.Key.(sarama.StringEncoder))
	rp.logger.LogMessageSuccess(key,
		msg.Topic,
		msg.Partition,
		msg.Offset,
		latencyMs)

	// Remove from failed messages if it was being retried
	rp.failedMsgMutex.Lock()
	delete(rp.failedMessages, key)
	rp.failedMsgMutex.Unlock()
}

// handleErrorEvents processes message delivery errors
func (rp *RobustProducer) handleErrorEvents(ctx context.Context) {
	defer rp.wg.Done()
	rp.logger.Info("❌ Starting error event handler...")

	for {
		select {
		case <-ctx.Done():
			rp.logger.Info("⏹️ Error handler stopped")
			return
		case errEvent := <-rp.errorEvents:
			rp.handleErroredMessage(errEvent)
		}
	}
}

// handleErroredMessage processes a failed message delivery
func (rp *RobustProducer) handleErroredMessage(errEvent *sarama.ProducerError) {
	key := string(errEvent.Msg.Key.(sarama.StringEncoder))

	// Classify the error
	errorInfo := errorhandling.ClassifyError(errEvent.Err)

	// Log the error with classification
	rp.logger.LogMessageError(key,
		errEvent.Msg.Topic,
		string(errorInfo.Type),
		errorInfo.Code,
		errEvent.Err,
		0) // Will be updated with actual retry count later

	// Update metrics based on error type
	if errorInfo.Retryable {
		rp.stats.IncrementTransientErrors()
		rp.metrics.RecordMessageError(errEvent.Msg.Topic, "transient", errorInfo.Code)
	} else {
		rp.stats.IncrementPermanentErrors()
		rp.metrics.RecordMessageError(errEvent.Msg.Topic, "permanent", errorInfo.Code)
	}

	// Handle failed message
	rp.failedMsgMutex.Lock()
	failure, exists := rp.failedMessages[key]
	if !exists {
		failure = errorhandling.NewMessageFailure(errEvent.Msg, errEvent.Err)
		rp.failedMessages[key] = failure
	} else {
		failure.Error = errEvent.Err
		failure.Classification = errorInfo
		failure.UpdateForRetry(rp.retryPolicy)
	}
	rp.failedMsgMutex.Unlock()

	// Decide on retry or DLQ
	if errorInfo.Retryable && failure.AttemptCount < rp.retryPolicy.MaxRetries {
		// Queue for retry
		select {
		case rp.retryQueue <- failure:
			rp.logger.Debug("Message queued for retry",
				"key", key,
				"attempts", failure.AttemptCount,
				"error", errEvent.Err)
		default:
			// Retry queue full, send to DLQ
			rp.sendToDeadLetterQueue(failure, "retry_queue_full")
		}
	} else {
		// Send to Dead Letter Queue
		reason := "max_retries_exceeded"
		if !errorInfo.Retryable {
			reason = "non_retryable_error"
		}
		rp.sendToDeadLetterQueue(failure, reason)
	}
}

// processRetryQueue handles message retries
func (rp *RobustProducer) processRetryQueue(ctx context.Context) {
	defer rp.wg.Done()
	rp.logger.Info("🔄 Starting retry queue processor...")

	for {
		select {
		case <-ctx.Done():
			rp.logger.Info("⏹️ Retry processor stopped")
			return
		case failure := <-rp.retryQueue:
			rp.processRetry(failure)
		}
	}
}

// processRetry attempts to retry a failed message
func (rp *RobustProducer) processRetry(failure *errorhandling.MessageFailure) {
	key := string(failure.Message.Key.(sarama.StringEncoder))

	// Calculate backoff delay
	backoffDelay := rp.retryPolicy.CalculateBackoff(failure.AttemptCount)

	rp.logger.Debug("Processing message retry",
		"key", key,
		"attempt", failure.AttemptCount,
		"backoff", backoffDelay)

	// Wait for backoff period
	time.Sleep(backoffDelay)

	// Update message timestamp for retry
	failure.Message.Timestamp = time.Now()
	failure.Message.Metadata = time.Now() // Reset start time for latency

	// Attempt retry
	select {
	case rp.producer.Input() <- failure.Message:
		rp.stats.IncrementRetried()
		// Record retry in metrics using error recording
		rp.metrics.RecordMessageError(failure.Message.Topic, "retry", "attempt")
		rp.logger.Debug("Message retry sent", "key", key)
	default:
		// Producer channel full, send to DLQ
		rp.sendToDeadLetterQueue(failure, "retry_channel_full")
	}
}

// sendToDeadLetterQueue sends a failed message to DLQ
func (rp *RobustProducer) sendToDeadLetterQueue(failure *errorhandling.MessageFailure, reason string) {
	key := string(failure.Message.Key.(sarama.StringEncoder))
	dlqTopic := rp.cfg.Kafka.Topic + ".dlq"

	// Create DLQ message with metadata
	dlqMessage := &sarama.ProducerMessage{
		Topic:     dlqTopic,
		Key:       failure.Message.Key,
		Value:     failure.Message.Value,
		Timestamp: time.Now(),
		Headers: []sarama.RecordHeader{
			{
				Key:   []byte("original-topic"),
				Value: []byte(failure.Message.Topic),
			},
			{
				Key:   []byte("failure-reason"),
				Value: []byte(reason),
			},
			{
				Key:   []byte("attempts"),
				Value: []byte(fmt.Sprintf("%d", failure.AttemptCount)),
			},
			{
				Key:   []byte("original-timestamp"),
				Value: []byte(failure.FirstAttempt.Format(time.RFC3339)),
			},
			{
				Key:   []byte("last-error"),
				Value: []byte(failure.Error.Error()),
			},
		},
	}

	// Send to DLQ
	select {
	case rp.dlqProducer.Input() <- dlqMessage:
		rp.stats.IncrementDeadLettered()
		// Record DLQ message using error recording
		rp.metrics.RecordMessageError(dlqTopic, "dead_letter", reason)
		rp.logger.LogDeadLetterQueue(key, failure.Message.Topic, dlqTopic, reason, failure.AttemptCount)

		// Remove from failed messages
		rp.failedMsgMutex.Lock()
		delete(rp.failedMessages, key)
		rp.failedMsgMutex.Unlock()
	default:
		rp.logger.Error("Failed to send message to DLQ - channel full",
			"key", key,
			"reason", reason)
		rp.metrics.RecordMessageError(dlqTopic, "dlq", "channel_full")
	}
}

// drainProducerChannels handles producer success/error events
func (rp *RobustProducer) drainProducerChannels(ctx context.Context) {
	defer rp.wg.Done()
	rp.logger.Info("📡 Starting producer channel drainer...")

	for {
		select {
		case <-ctx.Done():
			rp.logger.Info("⏹️ Channel drainer stopped")
			return
		case success := <-rp.producer.Successes():
			// Forward to buffered success channel
			select {
			case rp.successEvents <- success:
			default:
				rp.logger.Warn("Success events channel full, dropping event")
			}
		case err := <-rp.producer.Errors():
			// Forward to buffered error channel
			select {
			case rp.errorEvents <- err:
			default:
				rp.logger.Warn("Error events channel full, dropping event")
			}
		case success := <-rp.dlqProducer.Successes():
			// Log DLQ success
			key := string(success.Key.(sarama.StringEncoder))
			rp.logger.Debug("DLQ message delivered successfully", "key", key)
		case err := <-rp.dlqProducer.Errors():
			// Log DLQ failure
			key := string(err.Msg.Key.(sarama.StringEncoder))
			rp.logger.Error("DLQ message delivery failed",
				"key", key,
				"error", err.Err)
			rp.metrics.RecordMessageError(err.Msg.Topic, "dlq", "failure")
		}
	}
}

// startStatsReporting provides periodic status updates
func (rp *RobustProducer) startStatsReporting(ctx context.Context) {
	defer rp.wg.Done()
	rp.logger.Info("📊 Starting stats reporting...")

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			rp.logger.Info("⏹️ Stats reporting stopped")
			return
		case <-ticker.C:
			rp.reportCurrentStats()
		}
	}
}

// reportCurrentStats logs current production statistics
func (rp *RobustProducer) reportCurrentStats() {
	produced, _, _, transientErrs, permanentErrs, avgLatency, _ := rp.stats.GetStats()

	rp.logger.LogProducerHealth(
		produced,
		transientErrs+permanentErrs,
		(float64(produced)/float64(produced+transientErrs+permanentErrs))*100,
		avgLatency)

	// Update Prometheus gauges manually since we don't have specific methods
	rp.metrics.SetQueueSize(float64(len(rp.retryQueue)))

	// Log failed messages count
	rp.failedMsgMutex.RLock()
	failedCount := len(rp.failedMessages)
	rp.failedMsgMutex.RUnlock()

	if failedCount > 0 {
		rp.logger.Warn("Messages in retry queue", "count", failedCount)
	}
}

// PrintFinalStats displays comprehensive final statistics
func (rp *RobustProducer) PrintFinalStats() {
	produced, retried, dlq, transientErrs, permanentErrs, avgLatency, duration := rp.stats.GetStats()

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("📊 ROBUST PRODUCER FINAL STATISTICS")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("⏱️  Total Runtime: %v\n", duration.Round(time.Second))
	fmt.Printf("📤 Messages Produced: %d\n", produced)
	fmt.Printf("🔄 Messages Retried: %d\n", retried)
	fmt.Printf("💀 Messages Dead Lettered: %d\n", dlq)
	fmt.Printf("⚠️  Transient Errors: %d\n", transientErrs)
	fmt.Printf("❌ Permanent Errors: %d\n", permanentErrs)

	if produced > 0 {
		rate := float64(produced) / duration.Seconds()
		fmt.Printf("📈 Production Rate: %.2f msgs/sec\n", rate)
		fmt.Printf("⚡ Average Latency: %.2f ms\n", avgLatency)

		successRate := (float64(produced) / float64(produced+transientErrs+permanentErrs)) * 100
		fmt.Printf("✅ Success Rate: %.1f%%\n", successRate)
	}

	// Show pending failed messages
	rp.failedMsgMutex.RLock()
	pendingRetries := len(rp.failedMessages)
	rp.failedMsgMutex.RUnlock()

	if pendingRetries > 0 {
		fmt.Printf("⏳ Pending Retries: %d\n", pendingRetries)
	}

	fmt.Println(strings.Repeat("=", 80))
}

// Close gracefully shuts down the robust producer
func (rp *RobustProducer) Close() {
	rp.logger.Info("🔄 Closing robust producer...")

	// Cancel context to stop all goroutines
	rp.cancel()

	// Close producers
	if rp.producer != nil {
		rp.producer.AsyncClose()
	}
	if rp.dlqProducer != nil {
		rp.dlqProducer.AsyncClose()
	}

	rp.logger.Info("✅ Robust producer closed")
}

// WaitForShutdown waits for all goroutines to finish
func (rp *RobustProducer) WaitForShutdown() {
	rp.logger.Info("⏳ Waiting for all goroutines to finish...")
	rp.wg.Wait()
	rp.logger.Info("✅ All goroutines finished")
}
