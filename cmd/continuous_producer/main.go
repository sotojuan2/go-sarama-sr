package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
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

// ProducerStats tracks production metrics
type ProducerStats struct {
	mu               sync.RWMutex
	MessagesProduced int64
	ErrorsCount      int64
	StartTime        time.Time
	LastMessageTime  time.Time
}

func (ps *ProducerStats) IncrementProduced() {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.MessagesProduced++
	ps.LastMessageTime = time.Now()
}

func (ps *ProducerStats) IncrementErrors() {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.ErrorsCount++
}

func (ps *ProducerStats) GetStats() (int64, int64, time.Duration) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	duration := time.Since(ps.StartTime)
	return ps.MessagesProduced, ps.ErrorsCount, duration
}

// ContinuousProducer manages the continuous production loop
type ContinuousProducer struct {
	cfg        *config.Config
	srClient   *srClient.Client
	generator  *generator.ShoeGenerator
	serializer *protobuf.Serializer
	producer   sarama.AsyncProducer
	stats      *ProducerStats
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
}

func main() {
	log.Println("🚀 Starting Continuous Kafka Producer with Random Shoe Data Generation...")

	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Could not load .env file: %v", err)
	}

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("❌ Error loading config: %v", err)
	}

	log.Printf("📊 Configuration loaded:")
	log.Printf("   📡 Kafka Cluster: %s", cfg.Kafka.BootstrapServers)
	log.Printf("   📝 Topic: %s", cfg.Kafka.Topic)
	log.Printf("   ⏱️ Message Interval: %v", cfg.App.MessageInterval)
	log.Printf("   📈 Schema Registry: %s", cfg.SchemaRegistry.URL)

	// Create continuous producer
	continuousProducer, err := NewContinuousProducer(cfg)
	if err != nil {
		log.Fatalf("❌ Failed to create continuous producer: %v", err)
	}
	defer continuousProducer.Close()

	// Set up graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt signals
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// Start the continuous production loop
	log.Println("🔄 Starting continuous production loop...")
	log.Println("   Press Ctrl+C to stop gracefully")

	go continuousProducer.StartProduction(ctx)

	// Start stats reporting goroutine
	go continuousProducer.StartStatsReporting(ctx)

	// Wait for shutdown signal
	<-signalChan
	log.Println("\n⏹️ Shutdown signal received. Stopping production...")

	// Cancel context to stop production
	cancel()

	// Wait for graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	done := make(chan struct{})
	go func() {
		continuousProducer.WaitForShutdown()
		close(done)
	}()

	select {
	case <-done:
		log.Println("✅ Graceful shutdown completed")
	case <-shutdownCtx.Done():
		log.Println("⚠️ Shutdown timeout reached, forcing exit")
	}

	// Print final stats
	produced, errors, duration := continuousProducer.stats.GetStats()
	rate := float64(produced) / duration.Seconds()
	log.Printf("📊 Final Statistics:")
	log.Printf("   📤 Messages Produced: %d", produced)
	log.Printf("   ❌ Errors: %d", errors)
	log.Printf("   ⏱️ Runtime: %v", duration.Round(time.Second))
	log.Printf("   📈 Average Rate: %.2f messages/second", rate)
}

// NewContinuousProducer creates a new continuous producer instance
func NewContinuousProducer(cfg *config.Config) (*ContinuousProducer, error) {
	log.Println("🔧 Initializing continuous producer components...")

	// Create Schema Registry client
	srClientWrapper, err := srClient.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Schema Registry client: %w", err)
	}

	// Verify Schema Registry connectivity
	log.Println("🔍 Testing Schema Registry connectivity...")
	healthResult, err := srClientWrapper.ComprehensiveHealthCheck()
	if err != nil {
		return nil, fmt.Errorf("Schema Registry health check failed: %w", err)
	}
	log.Printf("✅ Schema Registry connectivity verified! Status: %+v", healthResult)

	// Initialize the random shoe generator
	shoeGen := generator.NewShoeGenerator()
	log.Println("✅ Random shoe generator initialized")

	// Create Protobuf serializer
	serializer, err := protobuf.NewSerializer(srClientWrapper.GetClient(), serde.ValueSerde, protobuf.NewSerializerConfig())
	if err != nil {
		return nil, fmt.Errorf("failed to create protobuf serializer: %w", err)
	}
	log.Println("✅ Protobuf serializer created")

	// Configure Sarama
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

	// Create async producer
	brokers := strings.Split(cfg.Kafka.BootstrapServers, ",")
	producer, err := sarama.NewAsyncProducer(brokers, saramaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create async producer: %w", err)
	}
	log.Println("✅ Kafka async producer created")

	ctx, cancel := context.WithCancel(context.Background())

	return &ContinuousProducer{
		cfg:        cfg,
		srClient:   srClientWrapper,
		generator:  shoeGen,
		serializer: serializer,
		producer:   producer,
		stats: &ProducerStats{
			StartTime: time.Now(),
		},
		ctx:    ctx,
		cancel: cancel,
	}, nil
}

// StartProduction starts the continuous production loop
func (cp *ContinuousProducer) StartProduction(ctx context.Context) {
	cp.wg.Add(1)
	defer cp.wg.Done()

	ticker := time.NewTicker(cp.cfg.App.MessageInterval)
	defer ticker.Stop()

	// Start success/error handling goroutines
	cp.wg.Add(2)
	go cp.handleSuccesses(ctx)
	go cp.handleErrors(ctx)

	log.Printf("🔄 Production loop started with interval: %v", cp.cfg.App.MessageInterval)

	for {
		select {
		case <-ctx.Done():
			log.Println("🛑 Production loop stopping...")
			return
		case <-ticker.C:
			// Generate random shoe
			shoe := cp.generator.GenerateRandomShoe()

			// Serialize shoe
			serializedData, err := cp.serializer.Serialize(cp.cfg.Kafka.Topic, shoe)
			if err != nil {
				log.Printf("❌ Serialization error: %v", err)
				cp.stats.IncrementErrors()
				continue
			}

			// Create message
			message := &sarama.ProducerMessage{
				Topic: cp.cfg.Kafka.Topic,
				Value: sarama.ByteEncoder(serializedData),
				Headers: []sarama.RecordHeader{
					{
						Key:   []byte("content-type"),
						Value: []byte("application/x-protobuf"),
					},
					{
						Key:   []byte("shoe-id"),
						Value: []byte(fmt.Sprintf("%d", shoe.Id)),
					},
					{
						Key:   []byte("shoe-brand"),
						Value: []byte(shoe.Brand),
					},
				},
			}

			// Send message asynchronously
			select {
			case cp.producer.Input() <- message:
				// Message queued successfully
			case <-ctx.Done():
				log.Println("🛑 Production loop stopping during message send...")
				return
			}
		}
	}
}

// handleSuccesses processes successful message deliveries
func (cp *ContinuousProducer) handleSuccesses(ctx context.Context) {
	defer cp.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case success := <-cp.producer.Successes():
			cp.stats.IncrementProduced()

			// Extract shoe ID from headers for logging
			shoeID := "unknown"
			for _, header := range success.Headers {
				if string(header.Key) == "shoe-id" {
					shoeID = string(header.Value)
					break
				}
			}

			if cp.cfg.App.LogLevel == "debug" {
				log.Printf("✅ Message produced: Shoe ID=%s, Partition=%d, Offset=%d",
					shoeID, success.Partition, success.Offset)
			}
		}
	}
}

// handleErrors processes message delivery errors
func (cp *ContinuousProducer) handleErrors(ctx context.Context) {
	defer cp.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case err := <-cp.producer.Errors():
			cp.stats.IncrementErrors()
			log.Printf("❌ Message production error: %v", err.Err)
		}
	}
}

// StartStatsReporting starts periodic stats reporting
func (cp *ContinuousProducer) StartStatsReporting(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second) // Report every 10 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			produced, errors, duration := cp.stats.GetStats()
			if produced > 0 {
				rate := float64(produced) / duration.Seconds()
				log.Printf("📊 Stats: Produced=%d, Errors=%d, Rate=%.2f msgs/sec, Runtime=%v",
					produced, errors, rate, duration.Round(time.Second))
			}
		}
	}
}

// Close closes the continuous producer
func (cp *ContinuousProducer) Close() error {
	log.Println("🔒 Closing continuous producer...")

	if cp.cancel != nil {
		cp.cancel()
	}

	if cp.producer != nil {
		err := cp.producer.Close()
		if err != nil {
			log.Printf("❌ Error closing producer: %v", err)
			return err
		}
	}

	log.Println("✅ Continuous producer closed successfully")
	return nil
}

// WaitForShutdown waits for all goroutines to finish
func (cp *ContinuousProducer) WaitForShutdown() {
	cp.wg.Wait()
}
