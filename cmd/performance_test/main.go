package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/IBM/sarama"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde/protobuf"
	"github.com/go-sarama-sr/producer/internal/config"
	"github.com/go-sarama-sr/producer/pkg/generator"
	srClient "github.com/go-sarama-sr/producer/pkg/schemaregistry"
	"github.com/joho/godotenv"
)

// PerformanceStats tracks detailed performance metrics
type PerformanceStats struct {
	StartTime        time.Time
	EndTime          time.Time
	MessagesProduced int64
	ErrorsCount      int64
	BytesProduced    int64
	MinLatency       time.Duration
	MaxLatency       time.Duration
	TotalLatency     time.Duration
	LatencyCount     int64
	mu               sync.RWMutex
}

func NewPerformanceStats() *PerformanceStats {
	return &PerformanceStats{
		StartTime:  time.Now(),
		MinLatency: time.Hour, // Start with a very high value
		MaxLatency: 0,
	}
}

func (ps *PerformanceStats) RecordSuccess(latency time.Duration, bytes int64) {
	atomic.AddInt64(&ps.MessagesProduced, 1)
	atomic.AddInt64(&ps.BytesProduced, bytes)

	ps.mu.Lock()
	ps.TotalLatency += latency
	ps.LatencyCount++
	if latency < ps.MinLatency {
		ps.MinLatency = latency
	}
	if latency > ps.MaxLatency {
		ps.MaxLatency = latency
	}
	ps.mu.Unlock()
}

func (ps *PerformanceStats) RecordError() {
	atomic.AddInt64(&ps.ErrorsCount, 1)
}

func (ps *PerformanceStats) Stop() {
	ps.EndTime = time.Now()
}

func (ps *PerformanceStats) PrintReport() {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	duration := ps.EndTime.Sub(ps.StartTime)
	messagesProduced := atomic.LoadInt64(&ps.MessagesProduced)
	errorsCount := atomic.LoadInt64(&ps.ErrorsCount)
	bytesProduced := atomic.LoadInt64(&ps.BytesProduced)

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("📊 PERFORMANCE TEST RESULTS")
	fmt.Println(strings.Repeat("=", 60))

	fmt.Printf("⏱️  Test Duration: %v\n", duration.Round(time.Millisecond))
	fmt.Printf("📤 Messages Produced: %d\n", messagesProduced)
	fmt.Printf("❌ Error Count: %d\n", errorsCount)
	fmt.Printf("📦 Total Bytes Produced: %s\n", formatBytes(bytesProduced))

	if duration > 0 {
		rate := float64(messagesProduced) / duration.Seconds()
		throughputMB := float64(bytesProduced) / (1024 * 1024) / duration.Seconds()
		fmt.Printf("📈 Average Rate: %.2f messages/second\n", rate)
		fmt.Printf("🚀 Throughput: %.2f MB/second\n", throughputMB)
	}

	if ps.LatencyCount > 0 {
		avgLatency := ps.TotalLatency / time.Duration(ps.LatencyCount)
		fmt.Printf("⚡ Min Latency: %v\n", ps.MinLatency.Round(time.Microsecond))
		fmt.Printf("⚡ Max Latency: %v\n", ps.MaxLatency.Round(time.Microsecond))
		fmt.Printf("⚡ Avg Latency: %v\n", avgLatency.Round(time.Microsecond))
	}

	if messagesProduced > 0 {
		successRate := float64(messagesProduced) / float64(messagesProduced+errorsCount) * 100
		fmt.Printf("✅ Success Rate: %.2f%%\n", successRate)
	}

	fmt.Println(strings.Repeat("=", 60))
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// PerformanceProducer optimized producer for performance testing
type PerformanceProducer struct {
	cfg        *config.Config
	srClient   *srClient.Client
	generator  *generator.ShoeGenerator
	serializer *protobuf.Serializer
	producer   sarama.AsyncProducer
	stats      *PerformanceStats
	wg         sync.WaitGroup
}

func main() {
	log.Println("🚀 Starting Kafka Performance Test...")

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Could not load .env file: %v", err)
	}

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("❌ Error loading config: %v", err)
	}

	// Create performance producer
	perfProducer, err := NewPerformanceProducer(cfg)
	if err != nil {
		log.Fatalf("❌ Failed to create performance producer: %v", err)
	}
	defer perfProducer.Close()

	// Set up graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// Performance test configurations
	testConfigs := []struct {
		name        string
		duration    time.Duration
		interval    time.Duration
		concurrency int
	}{
		{"🐌 Conservative Test", 30 * time.Second, 100 * time.Millisecond, 1},
		{"🏃 Moderate Test", 30 * time.Second, 50 * time.Millisecond, 2},
		{"🚀 Aggressive Test", 30 * time.Second, 10 * time.Millisecond, 4},
		{"💨 Maximum Test", 30 * time.Second, 1 * time.Millisecond, 8},
	}

	fmt.Println("\n📋 Performance Test Plan:")
	fmt.Println("   1. Conservative: 100ms interval, 1 goroutine")
	fmt.Println("   2. Moderate: 50ms interval, 2 goroutines")
	fmt.Println("   3. Aggressive: 10ms interval, 4 goroutines")
	fmt.Println("   4. Maximum: 1ms interval, 8 goroutines")
	fmt.Println("\n⏳ Each test runs for 30 seconds...")
	fmt.Println("💡 Press Ctrl+C to stop at any time\n")

	for i, testConfig := range testConfigs {
		select {
		case <-signalChan:
			log.Println("\n⏹️ Shutdown signal received. Stopping tests...")
			cancel()
			return
		default:
		}

		fmt.Printf("\n🔥 Starting Test %d: %s\n", i+1, testConfig.name)
		fmt.Printf("   Duration: %v, Interval: %v, Concurrency: %d\n",
			testConfig.duration, testConfig.interval, testConfig.concurrency)

		// Reset stats for this test
		perfProducer.stats = NewPerformanceStats()

		// Run the test
		testCtx, testCancel := context.WithTimeout(ctx, testConfig.duration)
		err := perfProducer.RunPerformanceTest(testCtx, testConfig.interval, testConfig.concurrency)
		testCancel()

		if err != nil {
			if ctx.Err() != nil {
				log.Println("⏹️ Test stopped due to shutdown signal")
				break
			}
			log.Printf("❌ Test failed: %v", err)
			continue
		}

		perfProducer.stats.Stop()
		perfProducer.stats.PrintReport()

		// Wait between tests (unless shutting down)
		if i < len(testConfigs)-1 {
			fmt.Println("\n⏸️ Waiting 5 seconds before next test...")
			select {
			case <-ctx.Done():
				return
			case <-time.After(5 * time.Second):
			}
		}
	}

	fmt.Println("\n🎉 Performance testing completed!")
}

func NewPerformanceProducer(cfg *config.Config) (*PerformanceProducer, error) {
	log.Println("🔧 Initializing performance producer...")

	// Create Schema Registry client
	srClientWrapper, err := srClient.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Schema Registry client: %w", err)
	}

	// Verify connectivity
	healthResult, err := srClientWrapper.ComprehensiveHealthCheck()
	if err != nil {
		return nil, fmt.Errorf("Schema Registry health check failed: %w", err)
	}
	log.Printf("✅ Schema Registry connected: %+v", healthResult)

	// Initialize generator
	shoeGen := generator.NewShoeGenerator()

	// Create serializer
	serializer, err := protobuf.NewSerializer(srClientWrapper.GetClient(), serde.ValueSerde, protobuf.NewSerializerConfig())
	if err != nil {
		return nil, fmt.Errorf("failed to create protobuf serializer: %w", err)
	}

	// Configure Sarama for high performance
	saramaConfig := sarama.NewConfig()
	saramaConfig.Producer.RequiredAcks = sarama.WaitForLocal // Faster than WaitForAll
	saramaConfig.Producer.Retry.Max = 3
	saramaConfig.Producer.Return.Successes = true
	saramaConfig.Producer.Return.Errors = true
	saramaConfig.Producer.Timeout = 10 * time.Second
	saramaConfig.Producer.Compression = sarama.CompressionSnappy // Enable compression
	saramaConfig.Producer.Flush.Frequency = 100 * time.Millisecond
	saramaConfig.Producer.Flush.Messages = 100
	saramaConfig.Producer.MaxMessageBytes = 10000000 // 10MB

	// SASL configuration
	saramaConfig.Net.SASL.Enable = true
	saramaConfig.Net.SASL.Mechanism = sarama.SASLTypePlaintext
	saramaConfig.Net.SASL.User = cfg.Kafka.APIKey
	saramaConfig.Net.SASL.Password = cfg.Kafka.APISecret
	saramaConfig.Net.TLS.Enable = true

	// Create producer
	brokers := strings.Split(cfg.Kafka.BootstrapServers, ",")
	producer, err := sarama.NewAsyncProducer(brokers, saramaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create async producer: %w", err)
	}

	return &PerformanceProducer{
		cfg:        cfg,
		srClient:   srClientWrapper,
		generator:  shoeGen,
		serializer: serializer,
		producer:   producer,
		stats:      NewPerformanceStats(),
	}, nil
}

func (pp *PerformanceProducer) RunPerformanceTest(ctx context.Context, interval time.Duration, concurrency int) error {
	log.Printf("🏁 Starting performance test: %d goroutines, %v interval", concurrency, interval)

	// Start success/error handlers
	pp.wg.Add(2)
	go pp.handleSuccesses(ctx)
	go pp.handleErrors(ctx)

	// Start producer goroutines
	for i := 0; i < concurrency; i++ {
		pp.wg.Add(1)
		go pp.produceMessages(ctx, interval, i)
	}

	// Start stats reporter
	pp.wg.Add(1)
	go pp.reportStats(ctx)

	// Wait for context cancellation
	<-ctx.Done()
	log.Println("⏹️ Test context cancelled, waiting for goroutines...")

	// Wait for all goroutines to finish
	pp.wg.Wait()
	log.Println("✅ All test goroutines finished")

	return nil
}

func (pp *PerformanceProducer) produceMessages(ctx context.Context, interval time.Duration, workerID int) {
	defer pp.wg.Done()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			startTime := time.Now()

			// Generate shoe
			shoe := pp.generator.GenerateRandomShoe()

			// Serialize
			serializedData, err := pp.serializer.Serialize(pp.cfg.Kafka.Topic, shoe)
			if err != nil {
				pp.stats.RecordError()
				continue
			}

			// Create message
			message := &sarama.ProducerMessage{
				Topic: pp.cfg.Kafka.Topic,
				Value: sarama.ByteEncoder(serializedData),
				Headers: []sarama.RecordHeader{
					{Key: []byte("content-type"), Value: []byte("application/x-protobuf")},
					{Key: []byte("worker-id"), Value: []byte(fmt.Sprintf("%d", workerID))},
					{Key: []byte("start-time"), Value: []byte(startTime.Format(time.RFC3339Nano))},
				},
			}

			// Send message
			select {
			case pp.producer.Input() <- message:
				// Message queued successfully
			case <-ctx.Done():
				return
			}
		}
	}
}

func (pp *PerformanceProducer) handleSuccesses(ctx context.Context) {
	defer pp.wg.Done()

	for {
		select {
		case <-ctx.Done():
			// Continue draining
			for success := range pp.producer.Successes() {
				pp.processSuccess(success)
			}
			return
		case success, ok := <-pp.producer.Successes():
			if !ok {
				return
			}
			pp.processSuccess(success)
		}
	}
}

func (pp *PerformanceProducer) processSuccess(success *sarama.ProducerMessage) {
	endTime := time.Now()

	// Calculate latency from headers
	var startTime time.Time
	for _, header := range success.Headers {
		if string(header.Key) == "start-time" {
			if t, err := time.Parse(time.RFC3339Nano, string(header.Value)); err == nil {
				startTime = t
				break
			}
		}
	}

	latency := endTime.Sub(startTime)
	messageSize := int64(len(success.Value.(sarama.ByteEncoder)))

	pp.stats.RecordSuccess(latency, messageSize)
}

func (pp *PerformanceProducer) handleErrors(ctx context.Context) {
	defer pp.wg.Done()

	for {
		select {
		case <-ctx.Done():
			// Continue draining
			for err := range pp.producer.Errors() {
				pp.stats.RecordError()
				log.Printf("❌ Producer error: %v", err.Err)
			}
			return
		case err, ok := <-pp.producer.Errors():
			if !ok {
				return
			}
			pp.stats.RecordError()
			log.Printf("❌ Producer error: %v", err.Err)
		}
	}
}

func (pp *PerformanceProducer) reportStats(ctx context.Context) {
	defer pp.wg.Done()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			produced := atomic.LoadInt64(&pp.stats.MessagesProduced)
			errors := atomic.LoadInt64(&pp.stats.ErrorsCount)
			duration := time.Since(pp.stats.StartTime)
			rate := float64(produced) / duration.Seconds()

			log.Printf("📊 Live: %d msgs, %d errors, %.1f msg/s", produced, errors, rate)
		}
	}
}

func (pp *PerformanceProducer) Close() error {
	log.Println("🔒 Closing performance producer...")

	if pp.producer != nil {
		pp.producer.AsyncClose()
	}

	return nil
}
