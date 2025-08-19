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

// SimpleStats tracks basic performance metrics
type SimpleStats struct {
	StartTime        time.Time
	MessagesProduced int64
	ErrorsCount      int64
	mu               sync.RWMutex
}

func NewSimpleStats() *SimpleStats {
	return &SimpleStats{
		StartTime: time.Now(),
	}
}

func (ss *SimpleStats) RecordSuccess() {
	atomic.AddInt64(&ss.MessagesProduced, 1)
}

func (ss *SimpleStats) RecordError() {
	atomic.AddInt64(&ss.ErrorsCount, 1)
}

func (ss *SimpleStats) GetRate() float64 {
	duration := time.Since(ss.StartTime).Seconds()
	if duration > 0 {
		return float64(atomic.LoadInt64(&ss.MessagesProduced)) / duration
	}
	return 0
}

func (ss *SimpleStats) PrintReport() {
	duration := time.Since(ss.StartTime)
	produced := atomic.LoadInt64(&ss.MessagesProduced)
	errors := atomic.LoadInt64(&ss.ErrorsCount)
	rate := ss.GetRate()

	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("📊 QUICK PERFORMANCE TEST RESULTS")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Printf("⏱️  Duration: %v\n", duration.Round(time.Second))
	fmt.Printf("📤 Messages: %d\n", produced)
	fmt.Printf("❌ Errors: %d\n", errors)
	fmt.Printf("📈 Rate: %.2f messages/second\n", rate)
	fmt.Println(strings.Repeat("=", 50))
}

// QuickProducer for fast testing
type QuickProducer struct {
	cfg        *config.Config
	srClient   *srClient.Client
	generator  *generator.ShoeGenerator
	serializer *protobuf.Serializer
	producer   sarama.AsyncProducer
	stats      *SimpleStats
	wg         sync.WaitGroup
}

func main() {
	log.Println("🚀 Quick Kafka Performance Test")
	log.Println("Press Ctrl+C to stop at any time")

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Could not load .env file: %v", err)
	}

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("❌ Error loading config: %v", err)
	}

	// Create quick producer
	quickProducer, err := NewQuickProducer(cfg)
	if err != nil {
		log.Fatalf("❌ Failed to create producer: %v", err)
	}
	defer quickProducer.Close()

	// Set up graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// Start production
	log.Println("🔄 Starting message production...")
	go quickProducer.StartProduction(ctx)

	// Start live stats reporting
	go quickProducer.StartLiveStats(ctx)

	// Wait for shutdown signal
	<-signalChan
	log.Println("\n⏹️ Stopping production...")
	cancel()

	// Wait for graceful shutdown
	quickProducer.WaitForShutdown()
	quickProducer.stats.PrintReport()
}

func NewQuickProducer(cfg *config.Config) (*QuickProducer, error) {
	log.Println("🔧 Setting up producer...")

	// Create Schema Registry client
	srClientWrapper, err := srClient.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Schema Registry client: %w", err)
	}

	// Quick health check
	_, err = srClientWrapper.ComprehensiveHealthCheck()
	if err != nil {
		return nil, fmt.Errorf("Schema Registry health check failed: %w", err)
	}
	log.Println("✅ Schema Registry connected")

	// Initialize generator
	shoeGen := generator.NewShoeGenerator()

	// Create serializer
	serializer, err := protobuf.NewSerializer(srClientWrapper.GetClient(), serde.ValueSerde, protobuf.NewSerializerConfig())
	if err != nil {
		return nil, fmt.Errorf("failed to create protobuf serializer: %w", err)
	}

	// Configure Sarama for performance
	saramaConfig := sarama.NewConfig()
	saramaConfig.Producer.RequiredAcks = sarama.WaitForLocal // Fast acknowledgment
	saramaConfig.Producer.Retry.Max = 2
	saramaConfig.Producer.Return.Successes = true
	saramaConfig.Producer.Return.Errors = true
	saramaConfig.Producer.Timeout = 5 * time.Second
	saramaConfig.Producer.Compression = sarama.CompressionSnappy

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

	return &QuickProducer{
		cfg:        cfg,
		srClient:   srClientWrapper,
		generator:  shoeGen,
		serializer: serializer,
		producer:   producer,
		stats:      NewSimpleStats(),
	}, nil
}

func (qp *QuickProducer) StartProduction(ctx context.Context) {
	qp.wg.Add(1)
	defer qp.wg.Done()

	// Start success/error handlers
	qp.wg.Add(2)
	go qp.handleSuccesses(ctx)
	go qp.handleErrors(ctx)

	// Production loop - aggressive rate for testing
	ticker := time.NewTicker(10 * time.Millisecond) // 100 msg/s target
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("🛑 Production stopping...")
			return
		case <-ticker.C:
			// Generate and serialize shoe
			shoe := qp.generator.GenerateRandomShoe()
			serializedData, err := qp.serializer.Serialize(qp.cfg.Kafka.Topic, shoe)
			if err != nil {
				qp.stats.RecordError()
				continue
			}

			// Create message
			message := &sarama.ProducerMessage{
				Topic: qp.cfg.Kafka.Topic,
				Value: sarama.ByteEncoder(serializedData),
				Headers: []sarama.RecordHeader{
					{Key: []byte("content-type"), Value: []byte("application/x-protobuf")},
					{Key: []byte("shoe-id"), Value: []byte(fmt.Sprintf("%d", shoe.Id))},
				},
			}

			// Send message
			select {
			case qp.producer.Input() <- message:
				// Queued successfully
			case <-ctx.Done():
				return
			}
		}
	}
}

func (qp *QuickProducer) handleSuccesses(ctx context.Context) {
	defer qp.wg.Done()

	for {
		select {
		case <-ctx.Done():
			// Drain remaining successes
			for range qp.producer.Successes() {
				qp.stats.RecordSuccess()
			}
			return
		case _, ok := <-qp.producer.Successes():
			if !ok {
				return
			}
			qp.stats.RecordSuccess()
		}
	}
}

func (qp *QuickProducer) handleErrors(ctx context.Context) {
	defer qp.wg.Done()

	for {
		select {
		case <-ctx.Done():
			// Drain remaining errors
			for err := range qp.producer.Errors() {
				qp.stats.RecordError()
				log.Printf("❌ Error: %v", err.Err)
			}
			return
		case err, ok := <-qp.producer.Errors():
			if !ok {
				return
			}
			qp.stats.RecordError()
			log.Printf("❌ Producer error: %v", err.Err)
		}
	}
}

func (qp *QuickProducer) StartLiveStats(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second) // Report every 2 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			produced := atomic.LoadInt64(&qp.stats.MessagesProduced)
			errors := atomic.LoadInt64(&qp.stats.ErrorsCount)
			rate := qp.stats.GetRate()

			if produced > 0 || errors > 0 {
				fmt.Printf("\r📊 Messages: %d | Errors: %d | Rate: %.1f msg/s | Duration: %v",
					produced, errors, rate, time.Since(qp.stats.StartTime).Round(time.Second))
			}
		}
	}
}

func (qp *QuickProducer) Close() error {
	if qp.producer != nil {
		qp.producer.AsyncClose()
	}
	return nil
}

func (qp *QuickProducer) WaitForShutdown() {
	qp.wg.Wait()
}
